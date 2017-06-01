// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/geodatalake/lambdas/scale"
	"github.com/satori/go.uuid"
)

func doc() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func array() *elastichelper.ArrayBuilder {
	return elastichelper.StartArray()
}

func produceJobTypeExtract() []byte {
	data := doc().
		AddKV("name", "detect-geo").
		AddKV("version", "1.0.0").
		AddKV("title", "Detect Geo").
		AddKV("description", "Extracts bounds and projection from geo files").
		AddKV("category", "testing").
		AddKV("author_name", "Steve_Ingram").
		AddKV("author_url", "http://www.example.com").
		AddKV("is_operational", true).
		AddKV("icon_code", "f279").
		AddKV("docker_privileged", false).
		AddKV("docker_image", "openwhere/scale-detector:dev").
		AddKV("priority", 230).
		AddKV("max_tries", 3).
		AddKV("cpus_required", 1.0).
		AddKV("mem_required", 1024.0).
		AddKV("disk_out_const_required", 0.0).
		AddKV("disk_out_mult_required", 0.0).
		Append("interface", doc().
			AddKV("version", "1.1").
			AddKV("command", "/opt/detect/detector").
			AddKV("command_arguments", "${extract_instructions} ${job_output_dir}").
			AddKV("shared_resources", []map[string]interface{}{}).
			AppendArray("output_data", array().
				Add(doc().
					AddKV("media_type", "application/json").
					AddKV("required", true).
					AddKV("type", "file").
					AddKV("name", "bounds_result"))).
			AppendArray("input_data", array().
				Add(doc().
					AppendArray("media_types", array().
						Add("application/json")).
					AddKV("required", true).
					AddKV("partial", true).
					AddKV("type", "file").
					AddKV("name", "extract_instructions")))).
		Append("error_mapping", doc().
			AddKV("version", "1.0").
			Append("exit_codes", doc().
				AddKV("10", "bad_num_input").
				AddKV("20", "open_input").
				AddKV("30", "read_input").
				AddKV("40", "marshal_failure").
				AddKV("50", "bad_s3_read").
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
}

func registerJobTypes(url, token string) {
	// Errors have to registered prior to job type ref'ing those errors
	createErrors(url, token)
	scale.RegisterJobType(url, token, produceJobTypeExtract())
}

//  Errors:
// 10 Bad number of input arguments
// 15 Bad Aws Session
// 20 Unable to open input
// 30 Unable to read input
// 40 Unable to marshal cluster request
// 50 Unable to read S3 bucket
// 70 Unknown cluster request
// 80 Unable to write to output
func main() {
	dev := flag.Bool("dev", false, "Development flag, interpret input as image file")
	jobType := flag.Bool("jt", false, "Output job type JSON to stdout")
	register := flag.String("register", "", "DC/OS Url, requires token")
	token := flag.String("token", "", "DC/OS token, required for register option")
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
		var cr geoindex.ClusterRequest
		if errJson := json.Unmarshal(b, &cr); errJson != nil {
			scale.WriteStderr(errJson.Error())
			os.Exit(40)
		}
		switch cr.RequestType {
		case geoindex.ExtractFileType:
			file := cr.File.File
			log.Println("Processing", cr.File.File)
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
				scale.WriteStderr(err.Error())
				os.Exit(15)
			}
			stream := bf.Stream(sess)
			resp, err := geoindex.DetectType(stream)
			if err != nil {
				log.Println("Not a geo")
			} else {
				data.Bounds = resp.Bounds
				data.Prj = resp.Prj
				data.Type = resp.Typ
				if resp.LastModified != "" {
					data.LastModified = resp.LastModified
				}
			}
			ended := time.Now().UTC()
			outName := path.Join(outdir, fmt.Sprintf("bounds_result_%s.json", uuid.NewV4().String()))
			scale.WriteJsonFile(outName, data)
			of := new(scale.OutputFile)
			of.Path = outName
			of.GeoMetadata = &scale.GeoMetadata{
				Started: started.Format(bucket.ISO8601FORMAT),
				Ended:   ended.Format(bucket.ISO8601FORMAT)}
			manifest := scale.FormatManifestFile("bounds_result", of, nil)
			scale.WriteJsonFile(path.Join(outdir, "results_manifest.json"), manifest)
			log.Println("DONE, exiting successful")
			os.Exit(0)
		default:
			scale.WriteStderr(fmt.Sprintf("Unknown request type %d", cr.RequestType))
			os.Exit(70)
		}
	} else {
		args := flag.Args()
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Println(err)
				os.Exit(10)
			}
			resp, err1 := geoindex.DetectType(f)
			f.Close()
			if err1 != nil {
				log.Println(err1)
				os.Exit(20)
			}
			log.Println(resp.Bounds, resp.Prj, resp.Typ, resp.LastModified)
		}
		os.Exit(0)
	}
	os.Exit(0)
}
