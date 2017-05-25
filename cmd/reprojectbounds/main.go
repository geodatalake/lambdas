// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strconv"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/geodatalake/lambdas/scale"
	"strings"
	"github.com/geodatalake/lambdas/proj4support"
	"github.com/ctessum/geom"
)

func WriteStderr(s string) {
	os.Stderr.Write([]byte(s))
}

func WriteJson(filePath string, objectToWrite interface{}) {
	if f, err := os.Create(filePath); err != nil {
		WriteStderr(fmt.Sprintf("Error writing %s: %v", filePath, err))
		os.Exit(20)
	} else {
		if jErr := json.NewEncoder(f).Encode(objectToWrite); jErr != nil {
			f.Close()
			WriteStderr(fmt.Sprintf("Error wrinting %s JSON: %v", filePath, jErr))
			os.Exit(30)
		}
		f.Close()
	}
}

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
	AddKV("60", "not_geo").
	AddKV("70", "bad_cluster_request")))
	b, err := json.MarshalIndent(data.Build(), "", "  ")

	if err != nil {
	WriteStderr(fmt.Sprintf("Error writing job type json: %v", err))
	os.Exit(-1)
	}
		return b
}

func createError(url, token string, data map[string]interface{}) {
	if data == nil {
		return
	}
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/errors/", url)
	var b []byte
	var err error
	if b, err = json.Marshal(data); err != nil {
		WriteStderr(err.Error())
		os.Exit(-1)
	}
	req, err1 := http.NewRequest("POST", urlStr, bytes.NewBuffer(b))
	if err1 != nil {
		WriteStderr(err1.Error())
		os.Exit(-1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	resp, err2 := client.Do(req)
	if err2 != nil {
		WriteStderr(err2.Error())
		os.Exit(-1)
	}
	resp.Body.Close()
	fmt.Println("Create New Error", data["name"], resp.Status)
}

func errorDoc(name, title, description string, existing map[string]int) map[string]interface{} {
	if _, ok := existing[name]; !ok {
		return doc().
		AddKV("name", name).
		AddKV("title", title).
		AddKV("description", description).
		AddKV("category", "ALGORITHM").Build()
	}
	return nil
}

type ExistingError struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	IsBuiltin    bool   `json:"is_builtin"`
	Created      string `json:"created"`
	LastModified string `json:"last_modified"`
}

type AllExistingErrors struct {
	Count   int             `json:"count"`
	Next    string          `json:"next,omitempty"`
	Prev    string          `json:"prev,omitempty"`
	Results []ExistingError `json:"results"`
}

func gatherExistingErrors(url, token string) map[string]int {

	retval := make(map[string]int)
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/errors/", url)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		WriteStderr(err.Error())
		os.Exit(-1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	fmt.Println("Url", urlStr, "Headers", req.Header)
	resp, err1 := client.Do(req)
	if err1 != nil {
		WriteStderr(err1.Error())
		os.Exit(-1)
	}
	b, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		WriteStderr(err2.Error())
		os.Exit(-1)
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		allErrors := new(AllExistingErrors)
	if errJson := json.Unmarshal(b, allErrors); errJson != nil {
		WriteStderr(errJson.Error())
		fmt.Println(" Response:", string(b))
		os.Exit(-1)
	}
	for _, existing := range allErrors.Results {
		retval[existing.Name] = existing.Id
		}
		return retval
	} else {
		fmt.Println(" Response:", string(b))
		os.Exit(-1)
	}
	return nil
}

func createErrors(url, token string) {
	existing := gatherExistingErrors(url, token)
	createError(url, token, errorDoc("bad_num_input", "Bad input cardinality", "Bad number of input arguments", existing))
	createError(url, token, errorDoc("open_input", "Failed to Open input", "Unable to open input", existing))
	createError(url, token, errorDoc("read_input", "Failed to Read input", "Unable to read input", existing))
	createError(url, token, errorDoc("marshal_failure", "Marshal JSON Failure", "Unable to marshal cluster request", existing))
	createError(url, token, errorDoc("bad_s3_read", "Failed S3 Bucket read", "Unable to read S3 bucket", existing))
	createError(url, token, errorDoc("not_geo", "Not a Geo File", "Unable to detect file type", existing))
	createError(url, token, errorDoc("bad_cluster_request", "Invalid Cluster Request", "Unknown cluster request", existing))
}

func validateJobType(url, token string) {

	// Errors have to registered prior to job type ref'ing those errors
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/job-types/validation/", url)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(produceJobType()))
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		WriteStderr(fmt.Sprintf("Error registering job type: %v", err))
		os.Exit(-1)
	}
	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			WriteStderr(err.Error())
			os.Exit(-1)
		}
		resp.Body.Close()
		fmt.Println(resp.Status, string(b))
	} else {
		fmt.Println("Response:", resp.Status)
	}
}

// curl -X POST -H "Authorization: token=${DCOS_TOKEN}" -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d @job-type.json "${DCOS_ROOT_URL}/service/scale/api/v5/job-types/"
func registerJobType(url, token string) {
	// Errors have to registered prior to job type ref'ing those errors
	createErrors(url, token)
	validateJobType(url, token)
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/job-types/", url)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(produceJobType()))
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		WriteStderr(fmt.Sprintf("Error registering job type: %v", err))
		os.Exit(-1)
	}
	if resp.StatusCode != 201 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			WriteStderr(err.Error())
			os.Exit(-1)
		}
		resp.Body.Close()
		fmt.Println(resp.Status, string(b))
	} else {
		fmt.Println("Response:", resp.Status)
	}
}

//  Errors:
// 10 Bad number of input arguments
// 20 Unable to open input
// 30 Unable to read input
// 40 Unable to marshal cluster request
// 50 Unable to read S3 bucket
// 60 Unable to detect file type
// 70 Unknown cluster request
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
		registerJobType(*register, *token)
		os.Exit(0)

	} else if *register != "" && *token == "" {
		WriteStderr("register requires token to also be specified")
		os.Exit(-1)
	} else if *token != "" && *register == "" {
		WriteStderr("token requires register to also be specified")
		os.Exit(-1)
	}

	if !*dev {
		started := time.Now().UTC()
		args := flag.Args()
		if len(args) != 2 {
			WriteStderr(fmt.Sprintf("Input arguments [%d] are not 2", len(args)))
			os.Exit(10)
		}
		input := args[0]
		outdir := args[1]
		f, err := os.Open(input)
		if err != nil {
			WriteStderr(fmt.Sprintf("Unable to open %s", input))
			os.Exit(20)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			WriteStderr(err.Error())
			os.Exit(30)
		}
		f.Close()
		var br scale.BoundsResult
		if errJson := json.Unmarshal(b, &br); errJson != nil {
			WriteStderr(errJson.Error())
			os.Exit(40)
		}
		outData := new(scale.OutputData)
		if len(br.Bounds) > 0  && len(br.Prj) > 0 {

			fmt.Println( br.Prj )
			var oldPrj = strings.TrimPrefix(br.Prj, "EPSG" )
			fmt.Println( oldPrj )
			oldPrj = strings.TrimSpace( oldPrj )
			br.Prj = "EPSG 4326"

			fmt.Println( oldPrj )
			jsonToParse := br.Bounds
			// First get offset to beginning of bounds array
			beginIndex := strings.Index(jsonToParse, "((")
			endIndex := strings.Index(jsonToParse, "))")
			featureType :=  jsonToParse[:beginIndex]
			substring := jsonToParse[ beginIndex + 2:endIndex]

			// Parse into pairs
			latLonPairs := strings.Split(substring, ",")

			var pts []geom.Point
			for _, row := range latLonPairs {

				elems := strings.Split(strings.TrimLeft(row," "), " ")

				if len(elems) >= 2 {

					fmt.Print("Lat: " + elems[0])
					fmt.Println("Lon: " + elems[1],)

					x,_ := strconv.ParseFloat(elems[0], 64 )
					y,_ := strconv.ParseFloat(elems[1], 64 )
					pts = append(pts, geom.Point{X: x, Y: y })

				}

			}

			fmt.Println( oldPrj )

			newPts := proj4support.ConvertPoints(oldPrj, pts)
			var newJsonPairs = ""
			for _, pt := range newPts {
				newJsonPairs = newJsonPairs + strconv.FormatFloat( pt.X, 'f', 6, 64) +","+  strconv.FormatFloat( pt.Y, 'f', 6, 64)
			}
			br.Bounds = featureType + "((" + newJsonPairs + "))"
			ended := time.Now().UTC()
			data := &scale.BoundsResult{Bounds: br.Bounds, Prj: br.Prj, Bucket: br.Bucket, Key: br.Key, Region: br.Region, LastModified: br.LastModified}
			outName := fmt.Sprintf("%s/bounds_result.json", outdir)
			WriteJson(outName, data)
			outData.Name = "bounds_result"
			outData.File = &scale.OutputFile{Path: outName, GeoMetadata: &scale.GeoMetadata{Started: started.Format(bucket.ISO8601FORMAT), Ended: ended.Format(bucket.ISO8601FORMAT)}}

		}
		outData.Name = "bounds_result"
		manifest := scale.FormatManifest([]*scale.OutputData{outData}, nil)
		WriteJson(fmt.Sprintf("%s/results_manifest.json", outdir), manifest)
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
