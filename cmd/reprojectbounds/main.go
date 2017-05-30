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
	"strconv"
	"strings"
	"time"

	"github.com/ctessum/geom"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/geodatalake/lambdas/proj4support"
	"github.com/geodatalake/lambdas/scale"
)

func doc() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func array() *elastichelper.ArrayBuilder {
	return elastichelper.StartArray()
}

func produceJobType() []byte {
	data := doc().
		AddKV("name", "reproject-boundary").
		AddKV("version", "1.0.0").
		AddKV("title", "Reproject Boundaries").
		AddKV("description", "Reproject bounds o").
		AddKV("category", "testing").
		AddKV("author_name", "Chris_Mangold").
		AddKV("author_url", "http://www.example.com").
		AddKV("is_operational", true).
		AddKV("icon_code", "f041").
		AddKV("docker_privileged", false).
		AddKV("docker_image", "openwhere/scale-reproject:dev").
		AddKV("priority", 230).
		AddKV("max_tries", 3).
		AddKV("cpus_required", 1.0).
		AddKV("mem_required", 1024.0).
		AddKV("disk_out_const_required", 0.0).
		AddKV("disk_out_mult_required", 0.0).
		Append("interface", doc().
			AddKV("version", "1.1").
			AddKV("command", "/opt/reproject/reprojectbounds").
			AddKV("command_arguments", "${bounds_result} ${job_output_dir}").
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
					AddKV("name", "bounds_result")))).
		Append("error_mapping", doc().
			AddKV("version", "1.0").
			Append("exit_codes", doc().
				AddKV("10", "bad_num_input").
				AddKV("20", "open_input").
				AddKV("30", "read_input").
				AddKV("40", "marshal_failure")))
	b, err := json.MarshalIndent(data.Build(), "", "  ")

	if err != nil {
		scale.WriteStderr(fmt.Sprintf("Error writing job type json: %v", err))
		os.Exit(-1)
	}
	return b
}

func createErrors(url, token string) {
	existing := scale.GatherExistingErrors(url, token)
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_num_input", "Bad input cardinality", "Bad number of input arguments", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("open_input", "Failed to Open input", "Unable to open input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("read_input", "Failed to Read input", "Unable to read input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("marshal_failure", "Marshal JSON Failure", "Unable to marshal cluster request", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_s3_read", "Failed S3 Bucket read", "Unable to read S3 bucket", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("not_geo", "Not a Geo File", "Unable to detect file type", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_cluster_request", "Invalid Cluster Request", "Unknown cluster request", existing))
}

//  Errors:
// 10 Bad number of input arguments
// 20 Unable to open input
// 30 Unable to read input
// 40 Unable to marshal cluster request
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
		fmt.Println(string(produceJobType()))
		os.Exit(0)
	}

	if *register != "" && *token != "" {
		scale.RegisterJobType(*register, *token, produceJobType())
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
		outData := new(scale.OutputData)
		if len(br.Bounds) > 0 && len(br.Prj) > 0 {
			fmt.Println(br.Prj)
			var oldPrj = strings.TrimPrefix(br.Prj, "EPSG")
			fmt.Println(oldPrj)
			oldPrj = strings.TrimSpace(oldPrj)
			br.Prj = "EPSG 4326"

			fmt.Println(oldPrj)
			jsonToParse := br.Bounds
			// First get offset to beginning of bounds array
			beginIndex := strings.Index(jsonToParse, "((")
			endIndex := strings.Index(jsonToParse, "))")
			featureType := jsonToParse[:beginIndex]
			substring := jsonToParse[beginIndex+2 : endIndex]

			// Parse into pairs
			latLonPairs := strings.Split(substring, ",")

			var pts []geom.Point
			for _, row := range latLonPairs {
				elems := strings.Split(strings.TrimLeft(row, " "), " ")
				if len(elems) >= 2 {
					fmt.Print("Lat: " + elems[0])
					fmt.Println("Lon: " + elems[1])

					x, _ := strconv.ParseFloat(elems[0], 64)
					y, _ := strconv.ParseFloat(elems[1], 64)
					pts = append(pts, geom.Point{X: x, Y: y})
				}
			}
			fmt.Println(oldPrj)
			newPts := proj4support.ConvertPoints(oldPrj, pts)
			var newJsonPairs = ""
			for _, pt := range newPts {
				newJsonPairs = newJsonPairs + strconv.FormatFloat(pt.X, 'f', 6, 64) + "," + strconv.FormatFloat(pt.Y, 'f', 6, 64)
			}
			br.Bounds = featureType + "((" + newJsonPairs + "))"
		}
		ended := time.Now().UTC()
		outName := fmt.Sprintf("%s/bounds_result.json", outdir)
		scale.WriteJsonFile(outName, br)
		outData.Name = "bounds_result"
		outData.File = &scale.OutputFile{Path: outName, GeoMetadata: &scale.GeoMetadata{Started: started.Format(bucket.ISO8601FORMAT), Ended: ended.Format(bucket.ISO8601FORMAT)}}

		outData.Name = "bounds_result"
		manifest := scale.FormatManifest([]*scale.OutputData{outData}, nil)
		scale.WriteJsonFile(fmt.Sprintf("%s/results_manifest.json", outdir), manifest)
		log.Println("Wrote", manifest.OutputData)
		os.Exit(0)
	} else {
		args := flag.Args()
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Println(err)
				os.Exit(10)
			}
			bounds, prj, geotype, err1 := geoindex.DetectType(f)
			f.Close()
			if err1 != nil {
				log.Println(err1)
				os.Exit(20)
			}
			log.Println(bounds, prj, geotype)
		}
		os.Exit(0)
	}
}
