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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
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
}

const (
	JScan    = "scan_bucket"
	JGroup   = "group_files"
	JProcess = "process_geo"

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

func (cr *ClusterRequest) Handle(specs HandlerSpec, ctx *runtime.Context) *JobSpec {
	js := new(JobSpec)
	js.Start = time.Now().UTC()
	switch cr.RequestType {
	case ScanBucket:
		mc := NewMonitorConn().WithRegion("us-west-2").WithFunctionArn(specs.GetMonitor())
		mc.InvokeIndexer(Enter, JScan, 1)
		js.Name = JScan
		log.Println("Initiating bucket scan", cr.Bucket)
		cq := new(ClusterQueue)
		items, err := scanBucket(cr, specs)
		js.Err = err
		if js.IsSuccess() {
			next, maxNext, _, _ := specs.GetNexts()
			cq.MaxNext = maxNext
			cq.Next = next
			cq.StartTime = time.Now().UTC().String()
			cq.ParentId = cr.Bucket.Bucket + "_" + cq.StartTime
			cq.Items = items
			masterCr := new(ClusterRequest)
			masterCr.RequestType = ClusterMaster
			masterCr.Master = cq
			masterCr.Id = cq.ParentId
			log.Println("Sending", len(items), "jobs to master")
			if _, err := AsyncCallNext(masterCr, next); err != nil {
				log.Println("Error invoking master lambda", err)
				js.Err = err
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
		} else {
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
	if cr.Master != nil {
		log.Println("Assuming master, starting processing a queue of", len(cr.Master.Items), "jobs", cr.Master.MaxNext, "jobs at a time")
		nq, timedout, complete := SendQueue(specs, cr.Master, ctx, false)
		for {
			if complete {
				st, err := time.Parse("2006-01-02 15:04:05.000000000 -0700 MST", cr.Master.StartTime)
				if err != nil {
					log.Println("Job", cr.Master.ParentId, "completed but time", cr.Master.StartTime, "could not be parsed", err)
				} else {
					dur := time.Now().UTC().Sub(st)
					log.Println("Job", cr.Master.ParentId, "Completed in", dur)
					ir := new(IndexerRequest)
					ir.Name = nq.ParentId
					ir.Duration = dur
					ir.RequestType = JobComplete
					mc := NewMonitorConn().WithRegion("us-west-2").WithFunctionArn(specs.GetMonitor())
					mc.Invoke(ir)
				}
				return nil
			}
			if timedout {
				if nq != nil {
					req := new(ClusterRequest)
					req.RequestType = ClusterMaster
					req.Master = nq
					AsyncCallNext(req, ctx.InvokedFunctionARN)
					return nil
				}
			} else if nq != nil {
				nq, timedout, complete = SendQueue(specs, nq, ctx, false)
			} else {
				return nil
			}
		}
	} else {
		log.Println("Error: No master property in ClusterRequest")
		return fmt.Errorf("No master property in ClusterRequest")
	}
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
			crOut.DirFiles = &DirRequest{Files: di.Keys}
			crOut.Id = cr.Id
			retval = append(retval, NewClusterResponse(crOut, cr.Id))
		}
	}

	log.Println("Processed", humanize.Comma(int64(count)), "items, size:", humanize.Bytes(uint64(size)), "Generated Requests:", len(retval))
	return retval, nil
}

func groupFiles(cr *ClusterRequest, specs HandlerSpec) ([]*ClusterResponse, error) {
	log.Println("Processing", cr)
	files, ok := Extract(cr.DirFiles)
	retval := []*ClusterResponse{}
	if ok {
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

func SendQueue(specs HandlerSpec, cq *ClusterQueue, ctx *runtime.Context, sendSqs bool) (*ClusterQueue, bool, bool) {
	numToSend := cq.MaxNext
	if numToSend > len(cq.Items) {
		numToSend = len(cq.Items)
	}
	numLeft := len(cq.Items) - numToSend
	if numLeft == 0 {
		numLeft = numToSend
	}

	// Allow the full job timeout + plus 10 seconds to spawn next sender
	var maxTimeout time.Duration
	counts := make([]int, 5)
	for _, cr := range cq.Items[:numToSend] {
		if maxTimeout > cr.Timeout {
			maxTimeout = cr.Timeout
		}
		counts[int(cr.Item.RequestType)]++
	}
	minDuration := maxTimeout + (10 * time.Second)
	if minDuration > ((4 * time.Minute) + (30 * time.Second)) {
		log.Println("Timeout can not exceed 4.5 minutes. Not possible to execute a job with timeout", maxTimeout.String())
		cq.Items = cq.Items[:0]
		return nil, false, true
	}

	jobDuration := time.Duration(ctx.RemainingTimeInMillis() * 1000 * 1000)

	if jobDuration < minDuration {
		return cq, true, len(cq.Items) == 0
	}
	jobTimeout := time.After(jobDuration - minDuration)

	ni := make([]*ClusterResponse, 0, numLeft)
	if len(cq.Items) > cq.MaxNext {
		for _, i := range cq.Items[numToSend:] {
			ni = append(ni, i)
		}
	}
	nextQ := cq.CloneWith(ni)

	indexer := NewMonitorConn().
		WithRegion("us-west-2").
		WithFunctionArn(specs.GetMonitor())

	indexer.InvokeIndexer(Enter, JScan, counts[0])
	indexer.InvokeIndexer(Enter, JProcess, counts[1])
	indexer.InvokeIndexer(Enter, JGroup, counts[2])
	var wait sync.WaitGroup
	wait.Add(numToSend)

	responses := make([]*ClusterQueue, numToSend)
	for idex, item := range cq.Items[0:numToSend] {
		go func(index int, cr *ClusterResponse) {
			if ok, io := sendJob(cr.Item, cq.Next, time.After(cr.Timeout)); !ok {
				nextQ.Items = append(nextQ.Items, cr)
			} else if io != nil {
				js := new(JobSpec)
				if err := json.Unmarshal(io.Payload, js); err != nil {
					log.Println("Error unmarshaling response payload", err, "Payload:", string(io.Payload))
				} else {
					responses[index] = cq.CloneWith(js.Items)
				}
			}
			wait.Done()
		}(idex, item)
	}
	waitDone := make(chan bool, 1)
	go func() {
		wait.Wait()
		waitDone <- true
	}()

	timedOut := false
	select {
	case <-waitDone:
		timedOut = false
		break
	case <-jobTimeout:
		timedOut = true
		break
	}

	indexer.InvokeIndexer(Leave, JScan, counts[0])
	indexer.InvokeIndexer(Leave, JProcess, counts[1])
	indexer.InvokeIndexer(Leave, JGroup, counts[2])

	for _, resp := range responses {
		if resp != nil && len(resp.Items) > 0 {
			nextQ.Items = append(nextQ.Items, resp.Items...)
		}
	}
	if len(nextQ.Items) > 0 {
		return nextQ, timedOut, false
	}
	return nextQ, false, true
}

func sendJob(cr *ClusterRequest, next string, timeout <-chan time.Time) (bool, *lambda.InvokeOutput) {
	resultChan, errChan := AlertCallNext(cr, next)
	for {
		select {
		case io := <-resultChan:
			return true, io
		case e := <-errChan:
			log.Println(e)
			return true, nil
		case <-timeout:
			return false, nil
		}
	}
}
