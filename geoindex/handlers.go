// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoindex

import (
	"context"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
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
	JScan    = "ReadBucket"
	JGroup   = "FilterFiles"
	JProcess = "DetectGeo"

	Contracts = "_contracts"
	Totalrun  = "_totalrun"
)

type JobSpec struct {
	Name     string    `json:"name"`
	Start    time.Time `json:"startTime"`
	End      time.Time `json:"endTime"`
	Duration string    `json:"duration"`
	Err      error     `json:"err"`
}

func (js *JobSpec) IsSuccess() bool {
	return js.Err == nil
}

type ContractTracker struct {
	client   *redis.Client
	provider *lambda.Lambda
	endpoint string
}

func NewContractTracker(client *redis.Client, provider *lambda.Lambda, endpoint string) *ContractTracker {
	return &ContractTracker{
		client:   client,
		provider: provider,
		endpoint: endpoint,
	}
}

func (ct *ContractTracker) Enter(name string) {
	if ct.client != nil {
		ct.client.Incr(name)
	} else {
		// TODO Send via SQS
		//ct.InvokeIndexer(Enter, name, 1)
	}
}

func (ct *ContractTracker) Leave(name string) {
	if ct.client != nil {
		ct.client.Decr(name)
	} else {
		// TODO Send via SQS
		// ct.InvokeIndexer(Leave, name, 1)
		ct.InvokeIndexer(Release, name, 1)
	}
}

func (ct *ContractTracker) Reserve(name string) bool {
	if ct.client != nil {
		cmd := ct.client.Decr(name + Contracts)
		if cmd.Val() >= 0 {
			return true
		} else {
			ct.client.Incr(name + Contracts)
			return false
		}
	} else {
		resp, _ := ct.InvokeIndexer(Reserve, name, 1)
		return resp.Success
	}
}

func (ct *ContractTracker) ReserveMany(name string, num int) int {
	if ct.client != nil {
		success := 0
		for i := num; i > 0; i-- {
			cmd := ct.client.Decr(name + Contracts)
			if cmd.Val() >= 0 {
				success++
			} else {
				ct.client.Incr(name + Contracts)
				break
			}
		}
		if name == JProcess {
			log.Println("ReserveMany(", num, ") reserved =", success)
			mySqs := NewSqsInstance().
				WithQueue("https://sqs.us-west-2.amazonaws.com/414519249282/process-geo-test-queue").
				WithRegion("us-west-2")
			if _, err := mySqs.Send(fmt.Sprintf("ReserveMany(%d) reserved=%d", num, success)); err != nil {
				log.Println("Error sending to SQS", err)
			}
		}
		return success
	} else {
		resp, _ := ct.InvokeIndexer(Reserve, name, num)
		if name == JProcess {
			log.Println("ReserveMany(", num, ") reserved =", resp.Num)
		}
		return resp.Num
	}
}
func (ct *ContractTracker) Release(name string) {
	if ct.client != nil {
		ct.client.Incr(name + Contracts)
	} else {
		ct.InvokeIndexer(Release, name, 1)
	}
}

func (ct *ContractTracker) Active() bool {
	return ct.client != nil || ct.endpoint != ""
}

type ContractFor struct {
	contract *ContractTracker
	jobName  string
}

func (ct *ContractFor) ReserveWait() bool {
	if ct.contract.Reserve(ct.jobName) {
		return true
	}
	tick := time.Tick(time.Second)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-tick:
			if ct.contract.Reserve(ct.jobName) {
				return true
			}
		case <-timeout:
			log.Println("Timed out waiting to send lambda")
			return false
		}
	}
}

func (ct *ContractFor) ReserveManyWait(num int) int {
	total := num
	if cnt := ct.contract.ReserveMany(ct.jobName, total); cnt == total {
		return num
	} else {
		total = total - cnt
	}
	tick := time.Tick(time.Second)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-tick:
			if cnt := ct.contract.ReserveMany(ct.jobName, total); cnt == total {
				return num
			} else {
				total = total - cnt
			}
		case <-timeout:
			return num - total
		}
	}
}

func NewContractFor(contract *ContractTracker, jobName string) *ContractFor {
	return &ContractFor{contract: contract, jobName: jobName}
}

func (cr *ClusterRequest) Handle(specs HandlerSpec) *JobSpec {
	contract := NewContractTracker(nil, nil, specs.GetMonitor())
	js := new(JobSpec)
	js.Start = time.Now().UTC()
	switch cr.RequestType {
	case ScanBucket:
		js.Name = JScan
		contract.Enter(JScan)
		log.Println("Initiating bucket scan", cr.Bucket)
		js.Err = scanBucket(cr, specs, NewContractFor(contract, JGroup))
		contract.Leave(JScan)
	case GroupFiles:
		js.Name = JGroup
		contract.Enter(JGroup)
		js.Err = groupFiles(cr, specs, NewContractFor(contract, JProcess))
		if cr.Contracted {
			contract.Leave(JGroup)
		}
	case ExtractFileType:
		js.Name = JProcess
		contract.Enter(JProcess)
		br, err := extractFile(cr, specs)
		if err != nil {
			js.Err = err
		} else {
			js.Err = index(br, specs)
		}
		if cr.Contracted {
			contract.Leave(JProcess)
		}
	}
	js.End = time.Now().UTC()
	js.Duration = js.End.Sub(js.Start).String()
	if js.Name == "" {
		js.Err = fmt.Errorf("Unhandled request %v", cr)
	}
	return js
}

func scanBucket(cr *ClusterRequest, specs HandlerSpec, contractFor *ContractFor) error {
	sess, err := scale.GetAwsSession()
	if err != nil {
		return err
	}
	var client *SqsInstance
	svc := s3.New(sess, aws.NewConfig().WithRegion(cr.Bucket.Region))
	root, err2 := bucket.ListBucketStructure(cr.Bucket.Region, cr.Bucket.Bucket, svc)
	if err2 != nil {
		log.Println("Error listing bucket contents", err2)
		return err2
	}
	q, r := specs.GetSqsOut()
	if !specs.UseLambda() && q != "" {
		client = NewSqsInstance().WithQueue(q).WithRegion(r)
	}
	iter := root.Iterate()
	count := 0
	chunkCount := 0
	next, maxNext, rate, _ := specs.GetNexts()
	chunk := NewChunkHandler(maxNext)
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
			chunkCount++
			if client != nil {
				_, sendErr := client.SendClusterRequest(crOut)
				if sendErr != nil {
					chunkCount--
					log.Println("Error queuing request", sendErr)
				}
			} else if specs.UseLambda() {
				if chunk.Add(crOut) {
					log.Println("Sending", chunk)
					chunk.Send(next, contractFor)
					if !contractFor.contract.Active() {
						time.Sleep(time.Duration(int64(rate)) * time.Second)
					}
				}
			}
		}
	}
	if specs.UseLambda() && !chunk.Empty() {
		chunk.Send(next, contractFor)
	}

	log.Println("Processed", humanize.Comma(int64(count)), "items, size:", humanize.Bytes(uint64(size)), "Generated Requests:", chunkCount)
	return nil
}

func groupFiles(cr *ClusterRequest, specs HandlerSpec, contractFor *ContractFor) error {
	var client *SqsInstance
	q, r := specs.GetSqsOut()
	if !specs.UseLambda() && q != "" {
		client = NewSqsInstance().WithQueue(q).WithRegion(r)
	}
	log.Println("Processing", cr)
	next, maxNext, rate, _ := specs.GetNexts()
	chunk := NewChunkHandler(maxNext)
	chunkCount := 0
	files, ok := Extract(cr.DirFiles)
	if ok {
		for _, ef := range files {
			clusterRequest := new(ClusterRequest)
			clusterRequest.RequestType = ExtractFileType
			clusterRequest.File = ef
			chunkCount++
			if client != nil {
				_, sendErr := client.SendClusterRequest(clusterRequest)
				if sendErr != nil {
					chunkCount--
					log.Println("Error queuing request", sendErr)
				}
			} else if specs.UseLambda() {
				if chunk.Add(clusterRequest) {
					chunk.Send(next, contractFor)
					if !contractFor.contract.Active() {
						time.Sleep(time.Duration(int64(rate)) * time.Second)
					}
				}
			}
		}
	}
	if specs.UseLambda() && !chunk.Empty() {
		chunk.Send(next, contractFor)
	}
	log.Println("Generated Requests:", chunkCount)
	return nil
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
