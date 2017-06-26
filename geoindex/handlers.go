// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoindex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/jobmanager"
	"github.com/geodatalake/lambdas/proj4support"
	"github.com/geodatalake/lambdas/scale"
	"github.com/go-redis/redis"
	elastic "gopkg.in/olivere/elastic.v5"
)

type HandlerSpec interface {
	EnableTraceLogging() bool
	GetESAuth() string
	FormatES() string
	GetESMethod() string
	GetESIndex() (string, string)
	GetSqsOut() (string, string)
	GetMonitor() string
	GetNexts() (string, int, int, int)
	UseLambda() bool
	UseStats() bool
	GetStats() *redis.Client
	DoIndex() bool
}

const (
	JScan    = "scan_bucket"
	JGroup   = "group_files"
	JProcess = "process_geo"

	JMaster  = "clusterMasters"
	JPending = "pendingJobs"

	Totalrun = "_totalrun"
)

type JobSpec struct {
	Name     string             `json:"name"`
	Start    time.Time          `json:"startTime"`
	End      time.Time          `json:"endTime"`
	Duration string             `json:"duration"`
	Err      error              `json:"err"`
	Items    []*ClusterResponse `json:"items"`
}

func (js *JobSpec) IsSuccess() bool {
	return js.Err == nil
}

type JobHelper struct {
	ctx *runtime.Context
}

func (jh *JobHelper) DecodeJobs(arr []interface{}) ([]interface{}, error) {
	retval := make([]interface{}, 0, len(arr))
	for _, obj := range arr {
		if m, ok := obj.(map[string]interface{}); ok {
			cr := new(ClusterResponse)
			if err1 := cr.Unmarshal(m); err1 == nil {
				retval = append(retval, cr)
			} else {
				return nil, fmt.Errorf("Error reading ClusterResponse: %v", err1)
			}
		} else {
			return nil, fmt.Errorf("Expected map[string]interface{} but got %s", reflect.TypeOf(obj).String())
		}
	}
	return retval, nil
}

// This is actually a JobSpec with embedded jobs
func (jh *JobHelper) UnmarshalJobs(data []byte) ([]interface{}, error) {
	log.Println("Unmarshaling", string(data))
	js := new(JobSpec)
	err := json.Unmarshal(data, js)
	if err != nil {
		log.Println("Error unmarshing", err)
		return nil, err
	}
	retval := make([]interface{}, 0, len(js.Items))
	for _, j := range js.Items {
		retval = append(retval, j)
	}
	return retval, err
}

func (jh *JobHelper) GetTimeout(job interface{}) time.Duration {
	if cr, ok := job.(*ClusterResponse); ok {
		return cr.Timeout
	} else {
		log.Println("WARN: expected job to be ClusterResponse pointer, it was", reflect.TypeOf(job).String())
	}
	return time.Minute
}

func (jh *JobHelper) GetType(job interface{}) string {
	if cr, ok := job.(*ClusterResponse); ok {
		switch cr.Item.RequestType {
		case ScanBucket:
			return JScan
		case ExtractFileType:
			return JProcess
		case GroupFiles:
			return JGroup
		}
	} else {
		log.Println("Expected job to be a ClusterResponse pointer, but it was", reflect.TypeOf(job).String())
	}
	return "UKNOWN"
}

func (jh *JobHelper) GetActualJob(job interface{}) interface{} {
	if cr, ok := job.(*ClusterResponse); ok {
		return cr.Item
	} else if req, good := job.(*ClusterRequest); good {
		return req
	} else {
		log.Println("Unknown job type", reflect.TypeOf(job).String())
	}
	return job
}

type OverallTimeout struct {
	ctx *runtime.Context
}

func (ot *OverallTimeout) Timeout() time.Duration {
	return time.Duration(ot.ctx.RemainingTimeInMillis() * 1000 * 1000)
}

func (cr *ClusterRequest) Handle(specs HandlerSpec, ctx *runtime.Context) *JobSpec {
	js := new(JobSpec)
	js.Start = time.Now().UTC()
	switch cr.RequestType {
	case ScanBucket:
		mc := NewMonitorConn().WithRegion("us-west-2").WithFunctionArn(specs.GetMonitor())
		mc.InvokeIndexer(Enter, JPending, 1)
		mc.InvokeIndexer(Enter, JScan, 1)
		js.Name = JScan
		log.Println("Initiating bucket scan", cr.Bucket)
		items, err := scanBucket(cr, specs)
		log.Println("items=", items)
		js.Err = err
		if js.IsSuccess() {
			startTime := time.Now().UTC().String()
			jobId := cr.Bucket.Bucket + "_" + startTime
			next, maxNext, _, _ := specs.GetNexts()
			mc.RegisterStart(JPending, len(items))
			parts := jobmanager.NewJobManager().CalcPackets(len(items), maxNext)
			log.Println("Job requires", parts, "parts")
			for i := 0; i < parts; i++ {
				job := jobmanager.NewClusterJob(jobId, startTime, i, parts-1)
				jp := new(jobmanager.JobPacket).
					WithNexts(next, maxNext).
					WithClusterJobs(job, nil)
				start, end := i*maxNext, (i*maxNext)+maxNext
				if end > len(items) {
					end = len(items)
				}
				log.Println("Processing start, end:", start, end)
				j := make([]interface{}, 0, end-start)
				for _, job := range items[start:end] {
					j = append(j, job)
				}
				jp.AddJobs(j)
				log.Println(jp)
				masterCr := new(ClusterRequest)
				masterCr.RequestType = ClusterMaster
				masterCr.Packet = jp
				masterCr.Id = jp.MyJob.Id
				log.Println("Sending", len(j), "jobs to master part", i, "of", parts-1)
				if _, err := AsyncCallNext(masterCr, next); err != nil {
					log.Println("Error invoking master lambda", err)
					js.Err = err
				}
			}
		}
		mc.InvokeIndexer(Leave, JScan, 1)
	case GroupFiles:
		js.Name = JGroup
		js.Items, js.Err = groupFiles(cr, specs)
	case ExtractFileType:
		js.Name = JProcess
		br, err := extractFile(cr, specs)
		if err != nil {
			js.Err = err
		} else if specs.DoIndex() {
			js.Err = index(br, specs)
		}
	case MineSQS:
		js.Name = "MineSQS"
		// TODO: Implement if needed
	case ClusterMaster:
		js.Name = "ClusterMaster"
		js.Err = MasterCluster(cr, specs, ctx)
	}
	js.End = time.Now().UTC()
	js.Duration = js.End.Sub(js.Start).String()
	if js.Name == "" {
		js.Err = fmt.Errorf("Unhandled request %v", cr)
	}
	return js
}

func MasterCluster(cr *ClusterRequest, specs HandlerSpec, ctx *runtime.Context) error {
	if cr.Packet != nil {
		mc := NewMonitorConn().WithRegion("us-west-2").WithFunctionArn(specs.GetMonitor())
		mc.InvokeIndexer(Enter, JPending, 1)
		mc.InvokeIndexer(Enter, JMaster, 1)
		defer mc.InvokeIndexer(Leave, JMaster, 1)

		manager := jobmanager.NewJobManager().
			WithSyncHelper(mc).
			WithDriver(NewLamabdaHelper().WithRegion("us-west-2")).
			WithJobHelper(&JobHelper{ctx: ctx}).
			WithJobTimeout(&OverallTimeout{ctx: ctx})
		if err := manager.RunMaster(cr.Packet, ctx.InvokedFunctionARN); err != nil {
			log.Println("Error: Executing RunMaster", err)
			return fmt.Errorf("Error executing RunMaster: %v", err)
		}
	} else {
		log.Println("Error: Packet is nil")
		return fmt.Errorf("Error: Packet is nil")
	}
	return nil
}

func scanBucket(cr *ClusterRequest, specs HandlerSpec) ([]*ClusterResponse, error) {
	sess, err := scale.GetAwsSession()
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess, aws.NewConfig().WithRegion(cr.Bucket.Region))
	root, err2 := bucket.ListBucketStructure(cr.Bucket.Region, cr.Bucket.Bucket, svc)
	if err2 != nil {
		log.Println("Error listing bucket contents", err2)
		return nil, err2
	}
	retval := make([]*ClusterResponse, 0, 128)
	iter := root.Iterate()
	count := 0
	size := int64(0)
	for {
		di, ok := iter.Next()
		if !ok {
			break
		}
		if len(di.Keys) > 0 {
			log.Println("Found", di.Name, "with", len(di.Keys), "files")
			count += len(di.Keys)
			size += di.Size
			crOut := new(ClusterRequest)
			crOut.RequestType = GroupFiles
			prefix, _ := path.Split(di.Keys[0].Key)
			crOut.DirFiles = &DirRequest{Bucket: cr.Bucket.Bucket, Region: cr.Bucket.Region, Prefix: prefix}
			crOut.Id = cr.Id
			retval = append(retval, NewClusterResponse(crOut, cr.Id))
		}
	}

	log.Println("Processed", humanize.Comma(int64(count)), "items, size:", humanize.Bytes(uint64(size)), "Generated Requests:", len(retval))
	return retval, nil
}

func groupFiles(cr *ClusterRequest, specs HandlerSpec) ([]*ClusterResponse, error) {
	log.Println("Processing", cr)
	sess, err := scale.GetAwsSession()
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess, aws.NewConfig().WithRegion(cr.DirFiles.Region))
	keys, err2 := bucket.ReadBucketDir(cr.DirFiles.Region, cr.DirFiles.Bucket, cr.DirFiles.Prefix, svc)
	if err2 != nil {
		return nil, err2
	}
	retval := []*ClusterResponse{}
	if files, ok := Extract(keys, cr.DirFiles.Prefix); ok {
		retval = make([]*ClusterResponse, 0, len(files))
		for _, ef := range files {
			clusterRequest := new(ClusterRequest)
			clusterRequest.RequestType = ExtractFileType
			clusterRequest.File = ef
			retval = append(retval, NewClusterResponse(clusterRequest, cr.Id))
		}
	}
	log.Println("Generated Requests:", len(retval))
	return retval, nil
}

func extractFile(cr *ClusterRequest, specs HandlerSpec) (*scale.BoundsResult, error) {
	file := cr.File.File
	bf := file.AsBucketFile()
	data := &scale.BoundsResult{Bucket: bf.Bucket, Key: bf.Key, Region: bf.Region, LastModified: bf.LastModified}
	if len(cr.File.Aux) > 0 {
		data.AuxFiles = make([]*scale.AuxResultFile, 0, len(cr.File.Aux))
		for _, f := range cr.File.Aux {
			data.AuxFiles = append(data.AuxFiles, &scale.AuxResultFile{Bucket: f.Bucket, Key: f.Key, Region: f.Region, LastModified: f.LastModified})
		}
	}
	if ext := path.Ext(file.Key); ext != "" {
		data.Extension = ext
	}
	sess, err := scale.GetAwsSession()
	if err != nil {
		return nil, err
	}
	stream := bf.Stream(sess)
	proj := &proj4support.ReProject{}
	resp, err := DetectType(stream, proj)
	if err != nil {
		if _, ok := err.(NotAGeo); !ok {
			return nil, err
		} else {
			log.Println("Not a geo")
		}
	} else {
		data.Bounds = resp.Bounds
		data.Prj = resp.Prj
		data.Type = resp.Typ
		if resp.LastModified != "" {
			data.LastModified = resp.LastModified
		}
	}
	return data, nil
}

type TraceLogger struct {
}

func (tl *TraceLogger) Printf(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

func doc() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func array() *elastichelper.ArrayBuilder {
	return elastichelper.StartArray()
}

func index(br *scale.BoundsResult, specs HandlerSpec) error {
	ctx := context.Background()

	url := specs.FormatES()
	log.Println("Using", url)
	logger := &TraceLogger{}
	opts := make([]elastic.ClientOptionFunc, 0, 16)
	opts = append(opts, elastic.SetURL(url))
	opts = append(opts, elastic.SetScheme(specs.GetESMethod()))
	opts = append(opts, elastic.SetHealthcheckTimeout(time.Second*10))
	opts = append(opts, elastic.SetHealthcheckTimeoutStartup(time.Second*10))
	opts = append(opts, elastic.SetSniff(false))
	opts = append(opts, elastic.SetInfoLog(logger))
	opts = append(opts, elastic.SetErrorLog(logger))
	if specs.EnableTraceLogging() {
		opts = append(opts, elastic.SetTraceLog(logger))
	}
	if specs.GetESAuth() != "" {
		splits := strings.Split(specs.GetESAuth(), ":")
		if len(splits) != 2 {
			return fmt.Errorf("Auth must be username:password format")

		}
		opts = append(opts, elastic.SetBasicAuth(splits[0], splits[1]))
	}

	client, err := elastic.NewClient(opts...)

	if err != nil {
		log.Println("Connection failed:", err)
		return err
	}

	nd := doc().
		AddKV("bucket", br.Bucket).
		AddKV("key", br.Key).
		AddKV("lastModified", br.LastModified).
		AddKV("region", br.Region)

	// "POLYGON ((%.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f))"
	// minX, maxY, maxX, maxY, maxX, minY
	if points, ok := elastichelper.ExtractPolygonPoints(br.Bounds); ok {
		var minX, minY, maxX, maxY float64
		allErrors := make(map[string]error)
		if len(points) < 6 {
			allErrors["short points"] = fmt.Errorf("Not enough points [%d] in bounds: %s", len(points), br.Bounds)
		} else {
			want := []int{0, 1, 2, 5}
			for i, w := range want {
				if val, err := strconv.ParseFloat(points[w], 64); err != nil {
					allErrors[points[w]] = err
				} else {
					switch i {
					case 0:
						minX = val
					case 1:
						maxY = val
					case 2:
						maxX = val
					case 3:
						minY = val
					}
				}
			}
		}
		if len(allErrors) > 0 {
			scale.WriteStderr(fmt.Sprintf("Parsing errors: %v", allErrors))
		} else {
			retval := elastichelper.MakeBboxClockwisePolygon(maxY, minX, minY, maxX)

			results := array()
			results.Add(retval)

			ndLocation := doc().AddKV("type", "polygon").
				AppendArray("coordinates", results).Build()

			nd.AddKV("bounds", br.Bounds).
				AddKV("projection", br.Prj).
				AddKV("type", br.Type).
				AddKV("location", ndLocation)
		}
	}

	indexToUse, indexTypeToUse := specs.GetESIndex()
	_, err = client.Index().Index(indexToUse).Type(indexTypeToUse).BodyJson(nd.Build()).Refresh("true").Do(ctx)

	if err != nil {
		return fmt.Errorf("Document Creation failed: %v", err)
	}
	return nil
}
