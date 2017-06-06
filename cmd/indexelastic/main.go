// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/scale"
	"github.com/satori/go.uuid"
	elastic "gopkg.in/olivere/elastic.v5"
)

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

func produceJobTypeExtract() []byte {
	data := doc().
		AddKV("name", "index-elastic").
		AddKV("version", "1.0.0").
		AddKV("title", "Index Elastic").
		AddKV("description", "Indexes into Elastic Search").
		AddKV("category", "testing").
		AddKV("author_name", "Steve_Ingram").
		AddKV("author_url", "http://www.example.com").
		AddKV("is_operational", true).
		AddKV("icon_code", "f07b").
		AddKV("docker_privileged", false).
		AddKV("docker_image", "openwhere/scale-index:dev").
		AddKV("priority", 230).
		AddKV("max_tries", 3).
		AddKV("cpus_required", 2.0).
		AddKV("mem_required", 2048.0).
		AddKV("disk_out_const_required", 0.0).
		AddKV("disk_out_mult_required", 0.0).
		Append("interface", doc().
			AddKV("version", "1.1").
			AddKV("command", "/opt/index/indexelastic").
			AddKV("command_arguments", "${index_request} ${job_output_dir}").
			AddKV("shared_resources", []map[string]interface{}{}).
			AppendArray("output_data", array().
				Add(doc().
					AddKV("media_type", "application/json").
					AddKV("required", true).
					AddKV("type", "file").
					AddKV("name", "index_record"))).
			AppendArray("input_data", array().
				Add(doc().
					AppendArray("media_types", array().
						Add("application/json")).
					AddKV("required", true).
					AddKV("partial", false).
					AddKV("type", "file").
					AddKV("name", "index_request")))).
		Append("error_mapping", doc().
			AddKV("version", "1.0").
			Append("exit_codes", doc().
				AddKV("10", "bad_num_input").
				AddKV("11", "es_conn").
				AddKV("20", "open_input").
				AddKV("30", "read_input").
				AddKV("33", "es_auth_format").
				AddKV("40", "marshal_failure").
				AddKV("50", "bad_s3_read").
				AddKV("55", "es_doc_write").
				AddKV("70", "bad_cluster_request").
				AddKV("80", "unable_write_output")))
	b, err := json.MarshalIndent(data.Build(), "", "  ")
	if err != nil {
		scale.WriteStderr(fmt.Sprintf("Error writing job type json: %v", err))
		os.Exit(-1)
	}
	return b
}

func createErrors(url, token string) {
	existing := scale.GatherExistingErrors(url, token)
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_session", "Bad AWS Session", "AWS Session failed to be created", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_num_input", "Bad input cardinality", "Bad number of input arguments", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("open_input", "Failed to Open input", "Unable to open input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("read_input", "Failed to Read input", "Unable to read input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("marshal_failure", "Marshal JSON Failure", "Unable to marshal cluster request", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_s3_read", "Failed S3 Bucket read", "Unable to read S3 bucket", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_cluster_request", "Invalid Cluster Request", "Unknown cluster request", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("unable_write_output", "Unable to write to output", "Unable to write to output", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("es_doc_write", "Unable to write elasticsearch", "Unable to write to ElasticSearch", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("es_conn", "Unable to connect to elasticsearch", "Unable to connect to ElasticSearch", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("es_auth_format", "ElasticSearch auths not formatted correctly", "ElasticSearch auths not formatted correctly", existing))
}

func registerJobTypes(url, token string) {
	// Errors have to registered prior to job type ref'ing those errors
	createErrors(url, token)
	scale.RegisterJobType(url, token, produceJobTypeExtract())
}

//  Errors:
// 10 Bad number of input arguments
// 11 ElasticSearch connection failed
// 15 Bad Aws Session
// 20 Unable to open input./
// 30 Unable to read input
// 33 Bad elastic search auths
// 40 Unable to marshal cluster request
// 50 Unable to read S3 bucket
// 55 Elastic Search Document creation failed
// 70 Unknown cluster request
// 80 Unable to write to output
func main() {
	dev := flag.Bool("dev", false, "Development flag, interpret input as image file")
	jobType := flag.Bool("jt", false, "Output job type JSON to stdout")
	register := flag.String("register", "", "DC/OS Url, requires token")
	token := flag.String("token", "", "DC/OS token, required for register option")
	host := flag.String("host", "34.223.221.59", "Elastic Search host instance")
	port := flag.String("port", "9200", "ElasticSearch port")
	auth := flag.String("auth", "elastic:changeme", "Elastic Search authentication to use")
	meth := flag.String("meth", "http", "Method to use")
	sniff := flag.Bool("sniff", false, "Enable sniffing")
	trace := flag.Bool("trace", false, "Enable trace logging")
	help := flag.Bool("h", false, "This help screen")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(-1)
	}

	if *jobType {
		fmt.Println(string(produceJobTypeExtract()))
		os.Exit(0)
	}

	if *register != "" && *token != "" {
		registerJobTypes(*register, *token)
		os.Exit(0)

	} else if *register != "" && *token == "" {
		scale.WriteStderr("register requires token to also be specified")
		os.Exit(-1)
	} else if *token != "" && *register == "" {
		scale.WriteStderr("token requires register to also be specified")
		os.Exit(-1)
	}

	if !*dev {
		started := time.Now().UTC()
		args := flag.Args()
		if len(args) != 2 {
			scale.WriteStderr(fmt.Sprintf("Input arguments [%d] are not 2", len(args)))
			os.Exit(10)
		}
		input := args[0]
		outdir := args[1]
		f, err := os.Open(input)
		if err != nil {
			scale.WriteStderr(fmt.Sprintf("Unable to open %s", input))
			os.Exit(20)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			scale.WriteStderr(err.Error())
			os.Exit(30)
		}
		f.Close()
		var br scale.BoundsResult
		if errJson := json.Unmarshal(b, &br); errJson != nil {
			scale.WriteStderr(errJson.Error())
			os.Exit(40)
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
				os.Exit(33)
			}
			opts = append(opts, elastic.SetBasicAuth(splits[0], splits[1]))
		}

		client, err := elastic.NewClient(opts...)

		if err != nil {
			log.Println("Connection failed:", err)
			os.Exit(11)
		}

		nd := doc().
			AddKV("bucket", br.Bucket).
			AddKV("key", br.Key).
			AddKV("lastModified", br.LastModified).
			AddKV("region", br.Region)

		// "POLYGON ((%.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f, %.7f %7f))"
		// minX, maxY, maxX, maxY, maxX, minY
		if strings.HasPrefix(br.Bounds, "POLYGON ((") {
			var minX, minY, maxX, maxY float64
			var err1, err2, err3, err4 error
			allErrors := make(map[string]error)
			points := strings.Split(br.Bounds[10:], ",")
			if len(points) < 7 {
				allErrors["short points"] = fmt.Errorf("Not enough points [%d] in bounds: %s", len(points), br.Bounds)
			} else {
				minX, err1 = strconv.ParseFloat(strings.TrimSpace(points[0]), 64)
				if err1 != nil {
					allErrors[strings.TrimSpace(points[0])] = err1
				}
				maxY, err2 = strconv.ParseFloat(strings.TrimSpace(points[1]), 64)
				if err2 != nil {
					allErrors[strings.TrimSpace(points[1])] = err2
				}
				maxX, err3 = strconv.ParseFloat(strings.TrimSpace(points[2]), 64)
				if err3 != nil {
					allErrors[strings.TrimSpace(points[2])] = err3
				}
				minY, err4 = strconv.ParseFloat(strings.TrimSpace(points[5]), 64)
				if err4 != nil {
					allErrors[strings.TrimSpace(points[5])] = err4
				}
			}
			if len(allErrors) > 0 {

				scale.WriteStderr(fmt.Sprintf("Parsing errors: %v", allErrors))

			} else {

				elastichelper.MakeEnvelope(maxY, minX, minY, maxX)

				retval := elastichelper.MakeBboxClockwisePolygon(maxY, minX, minY, maxX)
				
				results := array()
				results.Add( retval )

				ndLocation := doc().AddKV( "type", "polygon").
					AppendArray( "coordinates", results ).Build()

				nd.AddKV("bounds", br.Bounds).
					AddKV("projection", br.Prj).
					AddKV("type", br.Type).
					AddKV("location", ndLocation )

			}
		}

		_, err = client.Index().Index("sources").Type("source").BodyJson(nd.Build()).Refresh("true").Do(ctx)


		if err != nil {
			scale.WriteStderr(fmt.Sprintf("Document Creation failed: %v", err))
			os.Exit(55)
		}
		ended := time.Now().UTC()
		data := doc().AddKV("result", "Success").
			AddKV("bucket", br.Bucket).
			AddKV("key", br.Key).
			AddKV("region", br.Region).Build()
		outName := path.Join(outdir, fmt.Sprintf("index_record_%s.json", uuid.NewV4().String()))
		scale.WriteJsonFile(outName, data)
		of := new(scale.OutputFile)
		of.Path = outName
		of.GeoMetadata = &scale.GeoMetadata{
			Started: started.Format(bucket.ISO8601FORMAT),
			Ended:   ended.Format(bucket.ISO8601FORMAT)}
		manifest := scale.FormatManifestFile("index_record", of, nil)
		scale.WriteJsonFile(path.Join(outdir, "results_manifest.json"), manifest)
		os.Exit(0)
	} else {
		fmt.Println("No -dev option available")
		os.Exit(0)
	}
	os.Exit(0)
}
