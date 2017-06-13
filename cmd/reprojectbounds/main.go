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
	"strconv"
	"strings"
	"time"

	"github.com/geodatalake/geom"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/geodatalake/lambdas/proj4support"
	"github.com/geodatalake/lambdas/scale"
	"github.com/satori/go.uuid"
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
		AddKV("cpus_required", 2.0).
		AddKV("mem_required", 2048.0).
		AddKV("disk_out_const_required", 0.0).
		AddKV("disk_out_mult_required", 0.0).
		Append("interface", doc().
			AddKV("version", "1.1").
			AddKV("command", "/opt/reproject/reprojectbounds").
			AddKV("command_arguments", "${bounds_reproject} ${job_output_dir}").
			AddKV("shared_resources", []map[string]interface{}{}).
			AppendArray("output_data", array().
				Add(doc().
					AddKV("media_type", "application/json").
					AddKV("required", true).
					AddKV("type", "file").
					AddKV("name", "index_request"))).
			AppendArray("input_data", array().
				Add(doc().
					AppendArray("media_types", array().
						Add("application/json")).
					AddKV("required", true).
					AddKV("partial", false).
					AddKV("type", "file").
					AddKV("name", "bounds_reproject")))).
		Append("error_mapping", doc().
			AddKV("version", "1.0").
			Append("exit_codes", doc().
				AddKV("10", "bad_num_input").
				AddKV("20", "open_input").
				AddKV("30", "read_input").
				AddKV("40", "marshal_failure").
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
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_num_input", "Bad input cardinality", "Bad number of input arguments", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("open_input", "Failed to Open input", "Unable to open input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("read_input", "Failed to Read input", "Unable to read input", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("marshal_failure", "Marshal JSON Failure", "Unable to marshal cluster request", existing))
	scale.CreateScaleError(url, token, scale.ErrorDoc("unable_write_output", "Unable to write to output", "Unable to write to output", existing))
}

func buildNewReprojectionResponse( br scale.BoundsResult, binReadPath string ) scale.BoundsResult {


	fmt.Println(br.Prj)
	var oldPrj = br.Prj
	fmt.Println( "Old projection: " + oldPrj)

	oldPrj = strings.TrimSpace(oldPrj)
	br.Prj = "EPSG 4326"

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
	newPts,_ := proj4support.ConvertPoints(oldPrj, pts, binReadPath)
	var newJsonPairs = ""
	for _, pt := range newPts {
		newJsonPairs = newJsonPairs + strconv.FormatFloat(pt.X, 'f', 6, 64) + " " + strconv.FormatFloat(pt.Y, 'f', 6, 64) + ","
	}
	newJsonPairs = strings.TrimSuffix(newJsonPairs, ",")
	br.Bounds = featureType + "((" + newJsonPairs + "))"

	return br

}

//  Errors:
// 10 Bad number of input arguments
// 20 Unable to open input
// 30 Unable to read input
// 40 Unable to marshal bounds_result request
// 80 Unable to write to output
func main() {
	dev := flag.Bool("dev", false, "Development flag, interpret input as image file")
	jobType := flag.Bool("jt", false, "Output job type JSON to stdout")
	register := flag.String("register", "", "DC/OS Url, requires token")
	token := flag.String("token", "", "DC/OS token, required for register option")
	help := flag.Bool("h", false, "This help screen")
	binBuildSrcPath := flag.String("config", "", "Path to config files to genert binary files.")
	binBuildDestPath := flag.String("out", "", "Path to write generated binary files.")
	binReadPath := flag.String("binIn", "/opt/reproject/bins", "Path read binary binary files.")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(-1)
	}

	if *jobType {
		fmt.Println(string(produceJobType()))
		os.Exit(0)
	}

	if *binBuildSrcPath != "" {

		var destDirectory = ""
		if *binBuildDestPath != "" {
			destDirectory = *binBuildDestPath
		}
		proj4support.BuildMaps(*binBuildSrcPath, destDirectory)
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
		if len(br.Bounds) > 0 && len(br.Prj) > 0 {

			br = buildNewReprojectionResponse( br , *binReadPath )

		}
		ended := time.Now().UTC()
		outName := path.Join(outdir, fmt.Sprintf("index_request_%s.json", uuid.NewV4().String()))
		scale.WriteJsonFile(outName, br)
		of := new(scale.OutputFile)
		of.Path = outName
		of.GeoMetadata = &scale.GeoMetadata{
			Started: started.Format(bucket.ISO8601FORMAT),
			Ended:   ended.Format(bucket.ISO8601FORMAT)}
		manifest := scale.FormatManifestFile("index_request", of, nil)
		scale.WriteJsonFile(path.Join(outdir, "results_manifest.json"), manifest)
		os.Exit(0)
	} else {
		args := flag.Args()
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Println(err)
				os.Exit(10)
			}
			resp, err1 := geoindex.DetectType(f, nil)
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
