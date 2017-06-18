// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/go-redis/redis"
)

type TraceLogger struct {
}

func (tl *TraceLogger) Printf(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

//func main() {
//	os.Setenv("next", "arn:aws:lambda:us-west-2:414519249282:function:TestProcessGeo")
//	os.Setenv("maxNext", "2")
//	Handle(make(map[string]interface{}), nil)
//}

type RequestParams struct {
	Host         string
	Port         string
	Index        string
	IndexType    string
	Method       string
	TraceLog     bool
	ESauth       string
	SqsIn        string
	SqsOut       string
	SqsInRegion  string
	SqsOutRegion string
	Next         string
	MaxNext      int
	Rate         int
	Cap          int
	Stats        string
	StatsClient  *redis.Client
}

func loadFromEnv(name, deflt string) string {
	if retval, ok := os.LookupEnv(name); !ok {
		return deflt
	} else {
		return retval
	}
}

func NewParams() *RequestParams {
	params := new(RequestParams)
	params.Host = loadFromEnv("host", "34.223.221.59")
	params.Port = loadFromEnv("port", "9200")
	params.Index = loadFromEnv("index", "sources")
	params.IndexType = loadFromEnv("indexType", "source")
	params.Method = loadFromEnv("method", "http")
	params.TraceLog, _ = strconv.ParseBool(loadFromEnv("tracelog", "false"))
	params.ESauth = loadFromEnv("elasticAuth", "")
	params.SqsIn = loadFromEnv("sqsIn", "")
	params.SqsInRegion = loadFromEnv("sqsInRegion", "us-west-2")
	params.SqsOut = loadFromEnv("sqsOut", "")
	params.SqsOutRegion = loadFromEnv("sqsOutRegion", "us-west-2")
	params.Next = loadFromEnv("next", "")
	params.MaxNext, _ = strconv.Atoi(loadFromEnv("maxNext", "10"))
	params.Rate, _ = strconv.Atoi(loadFromEnv("rate", "5"))
	params.Cap, _ = strconv.Atoi(loadFromEnv("cap", "100"))
	params.Stats = loadFromEnv("stats", "")
	return params
}

func (rp *RequestParams) UseLambda() bool              { return rp.Next != "" }
func (rp *RequestParams) UseStats() bool               { return rp.Stats != "" }
func (rp *RequestParams) GetStats() *redis.Client      { return rp.StatsClient }
func (rp *RequestParams) EnableTraceLogging() bool     { return rp.TraceLog }
func (rp *RequestParams) GetESAuth() string            { return rp.ESauth }
func (rp *RequestParams) GetESMethod() string          { return rp.Method }
func (rp *RequestParams) GetESIndex() (string, string) { return rp.Index, rp.IndexType }
func (rp *RequestParams) GetSqsIn() (string, string)   { return rp.SqsIn, rp.SqsInRegion }
func (rp *RequestParams) GetSqsOut() (string, string)  { return rp.SqsOut, rp.SqsOutRegion }
func (rp *RequestParams) GetNexts() (string, int, int, int) {
	return rp.Next, rp.MaxNext, rp.Rate, rp.Cap
}
func (rp *RequestParams) FormatES() string {
	return fmt.Sprintf("%s://%s:%s", rp.Method, rp.Host, rp.Port)
}
func (rp *RequestParams) String() string {
	results := make([]string, 0, 16)
	results = append(results, "Host: "+rp.Host)
	results = append(results, "Port: "+rp.Port)
	results = append(results, "Index: "+rp.Index)
	results = append(results, "IndexType: "+rp.IndexType)
	results = append(results, "Method: "+rp.Method)
	results = append(results, fmt.Sprintf("TraceLog: %v", rp.TraceLog))
	results = append(results, "ESauth: "+rp.ESauth)
	results = append(results, "SqsIn: "+rp.SqsIn)
	results = append(results, "SqsInRegion: "+rp.SqsInRegion)
	results = append(results, "SqsOut: "+rp.SqsOut)
	results = append(results, "SqsOutRegion: "+rp.SqsOutRegion)
	results = append(results, "Next: "+rp.Next)
	results = append(results, fmt.Sprintf("MaxNext: %d", rp.MaxNext))
	results = append(results, fmt.Sprintf("Rate: %d", rp.Rate))
	results = append(results, fmt.Sprintf("Cap: %d", rp.Cap))
	results = append(results, "Stats: "+rp.Stats)
	return strings.Join(results, ", ")
}

func Handle(evt interface{}, ctx *runtime.Context) (interface{}, error) {
	params := NewParams()
	log.Println("Params", params)
	if params.UseStats() {
		params.StatsClient = redis.NewClient(&redis.Options{
			Addr:     params.Stats,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	}
	if m, ok := evt.(map[string]interface{}); ok {
		req := new(geoindex.ClusterRequest)
		if params.UseLambda() {
			if unErr := req.Unmarshal(m); unErr != nil {
				return make(map[string]string), unErr
			}
		} else {
			if unErr := json.Unmarshal([]byte(m["Body"].(string)), req); unErr != nil {
				return make(map[string]string), unErr
			}
		}
		js := req.Handle(params)
		if js.IsSuccess() {
			if !params.UseLambda() {
				q, r := params.GetSqsIn()
				if q != "" {
					client := geoindex.NewSqsInstance().WithQueue(q).WithRegion(r)
					if rc, ok := m["ReceiptHandle"]; ok {
						client.Delete(rc.(string))
					}
				}
			}
			return js, nil
		} else {
			return nil, js.Err
		}
	} else {
		return nil, fmt.Errorf("evt is not of type map[string]interface{}")
	}
}
