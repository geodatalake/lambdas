// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/elastichelper"
	elastic "gopkg.in/olivere/elastic.v5"
)

type TraceLogger struct {
}

func (tl *TraceLogger) Printf(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

//func main() {}

func Handle(evt interface{}, ctx *runtime.Context) (string, error) {
	var host, port, index, meth string
	var ok bool

	log.Printf("Log stream name: %s\n", ctx.LogStreamName)
	log.Printf("Log group name: %s\n", ctx.LogGroupName)
	log.Printf("Request ID: %s\n", ctx.AWSRequestID)
	log.Printf("Mem. limits(MB): %d\n", ctx.MemoryLimitInMB)

	log.Println("ctx", ctx)
	if ctx.ClientContext != nil && ctx.ClientContext.Environment != nil {
		env := ctx.ClientContext.Environment
		host, ok = env["host"]
		if !ok {
			host = "34.223.221.59"
		}
		port, ok = env["port"]
		if !ok {
			port = "9200"
		}
		index, ok = env["index"]
		if !ok {
			index = "sources"
		}
		meth, ok = env["method"]
		if !ok {
			meth = "http"
		}
	} else {
		host = "34.223.221.59"
		port = "9200"
		index = "sources"
		meth = "http"
	}

	myCtx := context.Background()

	url := fmt.Sprintf("%s://%s:%s", meth, host, port)
	log.Println("Using", url, "sniff is false")
	logger := &TraceLogger{}
	opts := make([]elastic.ClientOptionFunc, 0, 16)
	opts = append(opts, elastic.SetURL(url))
	opts = append(opts, elastic.SetScheme(meth))
	opts = append(opts, elastic.SetHealthcheckTimeout(time.Second*10))
	opts = append(opts, elastic.SetHealthcheckTimeoutStartup(time.Second*10))
	opts = append(opts, elastic.SetSniff(false))
	opts = append(opts, elastic.SetInfoLog(logger))
	opts = append(opts, elastic.SetErrorLog(logger))

	client, err := elastic.NewClient(opts...)

	if err != nil {
		log.Println("Connection failed:", err)
		return "Connection to ElasticSearch failed", err
	}

	bbox := elastichelper.NewBboxShapeQuery("location", 35.0, 51.0, 32.0, 54.0)
	q := elastic.NewBoolQuery()
	q = q.Must(elastic.NewMatchAllQuery())
	q = q.Filter(bbox)
	searchResult, err := client.Search(index).Query(q).Do(myCtx)
	if err != nil {
		log.Println("Search failed:", err)
		return "Search Failed", err
	}
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	for _, hit := range searchResult.Hits.Hits {
		var m interface{}
		json.Unmarshal(*hit.Source, &m)
		log.Println(m)
	}

	return "Hello World", nil
}
