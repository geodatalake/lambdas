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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/geodatalake/lambdas/elastichelper"
	awsauth "github.com/smartystreets/go-aws-auth"
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

type AWSSigningTransport struct {
	HTTPClient  *http.Client
	Credentials awsauth.Credentials
}

// RoundTrip implementation
func (a AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return a.HTTPClient.Do(awsauth.Sign4(req, a.Credentials))
}

func main() {
	host := flag.String("host", "34.223.221.59", "Elastic Search Host")
	port := flag.String("port", "9200", "Elastic Search Port")
	auth := flag.String("auth", "", "user:password for elastic basic auth")
	aws := flag.Bool("aws", false, "Enable aws signing client")
	meth := flag.String("meth", "http", "Method to use")
	sniff := flag.Bool("sniff", false, "Enable sniffing")
	trace := flag.Bool("trace", false, "Enable trace logging")
	createIndex := flag.Bool("createindex", false, "Force index mapping creation, does not delete it")
	dropIndex := flag.Bool("dropindex", false, "Force index deletion")
	help := flag.Bool("h", false, "Shows help info")
	flag.Parse()

	if *help {
		fmt.Printf("\nelastictest version 1.0.0\n\n")
		flag.Usage()
		os.Exit(10)
	}

	ctx := context.Background()

	url := fmt.Sprintf("%s://%s:%s", *meth, *host, *port)
	log.Println("Using", url, "sniff is", *sniff)
	logger := &TraceLogger{}
	opts := make([]elastic.ClientOptionFunc, 0, 16)
	opts = append(opts, elastic.SetURL(url))
	opts = append(opts, elastic.SetScheme(*meth))
	opts = append(opts, elastic.SetHealthcheckTimeout(time.Second*10))
	opts = append(opts, elastic.SetHealthcheckTimeoutStartup(time.Second*10))
	opts = append(opts, elastic.SetSniff(*sniff))
	opts = append(opts, elastic.SetInfoLog(logger))
	opts = append(opts, elastic.SetErrorLog(logger))
	if *trace {
		opts = append(opts, elastic.SetTraceLog(logger))
	}
	if *auth != "" {
		splits := strings.Split(*auth, ":")
		if len(splits) != 2 {
			log.Println("Auth must be username:password format")
			os.Exit(10)
		}
		opts = append(opts, elastic.SetBasicAuth(splits[0], splits[1]))
	}
	if *aws {
		signingTransport := AWSSigningTransport{
			Credentials: awsauth.Credentials{
				AccessKeyID:     os.Getenv("AWS_ACCESS_KEY"),
				SecretAccessKey: os.Getenv("AWS_SECRET_KEY"),
			},
			HTTPClient: http.DefaultClient,
		}
		signingClient := &http.Client{Transport: http.RoundTripper(signingTransport)}
		opts = append(opts, elastic.SetHttpClient(signingClient))
		if *meth == "http" {
			log.Println("Warning: AWS endpoints are usually https, use -meth https")
		}
	}
	client, err := elastic.NewClient(opts...)

	if err != nil {
		log.Println("Connection failed:", err)
		os.Exit(10)
	}

	if *dropIndex {
		client.DeleteIndex("sources").Do(ctx)
		os.Exit(0)
	}

	mapping := nd().
		Append("source", nd().
			Append("properties", nd().
				Append("name", nd().
					AddKV("type", "text")).
				Append("bucket", nd().
					AddKV("type", "text")).
				Append("key", nd().
					AddKV("type", "text")).
				Append("lastModified", nd().
					AddKV("type", "date")).
				Append("type", nd().
					AddKV("type", "text")).
				Append("bounds", nd().
					AddKV("type", "text")).
				Append("location", elastichelper.NewLocationMapping())))

	index := nd().Append("mappings", mapping)

	_, err = client.CreateIndex("sources").BodyJson(index.Build()).Do(ctx)
	if err != nil {
		log.Println("Index Creation failed:", err)
		os.Exit(10)
	}

	if *createIndex {
		fmt.Println("Index sources Created")
		os.Exit(0)
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
