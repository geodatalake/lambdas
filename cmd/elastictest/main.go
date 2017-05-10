// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/geodatalake/lambdas/elastichelper"
	elastic "gopkg.in/olivere/elastic.v5"
)

type TraceLogger struct {
}

func (tl *TraceLogger) Printf(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

func nd() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func main() {
	host := flag.String("host", "localhost", "Elastic Search Host")
	port := flag.String("port", "9200", "Elastic Search Port")
	auth := flag.String("auth", "", "user:password for elastic basic auth")
	sniff := flag.Bool("sniff", false, "Enable sniffing")
	trace := flag.Bool("trace", false, "Enable trace logging")
	help := flag.Bool("h", false, "Shows help info")
	flag.Parse()

	if *help {
		fmt.Println("\nelastictest version 1.0.0\n")
		flag.Usage()
		os.Exit(10)
	}

	ctx := context.Background()

	url := fmt.Sprintf("http://%s:%s", *host, *port)
	log.Println("Using", url, "sniff is", *sniff)
	logger := &TraceLogger{}
	opts := make([]elastic.ClientOptionFunc, 0, 16)
	opts = append(opts, elastic.SetURL(url))
	opts = append(opts, elastic.SetHealthcheckTimeout(time.Second*10))
	opts = append(opts, elastic.SetHealthcheckTimeoutStartup(time.Second*10))
	opts = append(opts, elastic.SetSniff(*sniff))
	opts = append(opts, elastic.SetInfoLog(logger))
	opts = append(opts, elastic.SetErrorLog(logger))
	if *trace {
		opts = append(opts, elastic.SetTraceLog(logger))
	}
	opts = append(opts, elastic.SetScheme("http"))
	if *auth != "" {
		splits := strings.Split(*auth, ":")
		if len(splits) != 2 {
			log.Println("Auth must be username:password format")
			os.Exit(10)
		}
		opts = append(opts, elastic.SetBasicAuth(splits[0], splits[1]))
	}
	client, err := elastic.NewClient(opts...)

	if err != nil {
		log.Println("Connection failed:", err)
		os.Exit(10)
	}
	mapping := nd().
		Append("source", nd().
			Append("properties", nd().
				Append("name", nd().
					AddKV("type", "text")).
				Append("location", elastichelper.NewLocationMapping())))

	index := nd().Append("mappings", mapping)

	_, err = client.CreateIndex("sources").BodyJson(index.Build()).Do(ctx)
	if err != nil {
		log.Println("Index Creation failed:", err)
		os.Exit(10)
	}

	doc := nd().
		AddKV("name", "foobar_test").
		Append("location", elastichelper.NewEnvelope(34.0, 52.0, 33.0, 53.0))

	_, err = client.Index().Index("sources").Type("source").BodyJson(doc.Build()).Refresh("true").Do(ctx)
	if err != nil {
		log.Println("Document Creation failed", err)
		os.Exit(10)
	}
	doc2 := nd().
		AddKV("name", "foobar_test2").
		Append("location", elastichelper.MakePoint(33.5, 52.5))
	_, err = client.Index().Index("sources").Type("source").BodyJson(doc2.Build()).Refresh("true").Do(ctx)
	if err != nil {
		log.Println("Document Creation failed", err)
		os.Exit(10)
	}

	bbox := elastichelper.NewBboxShapeQuery("location", 35.0, 51.0, 32.0, 54.0)
	q := elastic.NewBoolQuery()
	q = q.Must(elastic.NewMatchAllQuery())
	q = q.Filter(bbox)
	searchResult, err := client.Search("sources").Query(q).Do(ctx)
	if err != nil {
		log.Println("Search failed:", err)
		os.Exit(10)
	}
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	for _, hit := range searchResult.Hits.Hits {
		var m interface{}
		json.Unmarshal(*hit.Source, &m)
		fmt.Println(m)
	}
	client.DeleteIndex("sources").Do(ctx)

	log.Println("Success")
	os.Exit(0)
}
