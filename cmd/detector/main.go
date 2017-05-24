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

	"github.com/dustin/go-humanize"
	"github.com/geodatalake/lambdas/bucket"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/geodatalake/lambdas/scale"
	"github.com/satori/go.uuid"
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
		AddKV("name", "detect-geo").
		AddKV("version", "1.0.0").
		AddKV("title", "Detect Geo").
		AddKV("description", "Extracts bounds and projection from geo files").
		AddKV("category", "testing").
		AddKV("author_name", "Steve_Ingram").
		AddKV("author_url", "http://www.example.com").
		AddKV("is_operational", true).
		AddKV("icon_code", "f02b").
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
		var cr geoindex.ClusterRequest
		if errJson := json.Unmarshal(b, &cr); errJson != nil {
			WriteStderr(errJson.Error())
			os.Exit(40)
		}
		outData := new(scale.OutputData)
		switch cr.RequestType {
		case geoindex.ScanBucket:
			root, err2 := bucket.ListBucketStructure(cr.Bucket.Region, cr.Bucket.Bucket)
			if err2 != nil {
				WriteStderr(err2.Error())
				os.Exit(50)
			}
			iter := root.Iterate()
			count := 0
			size := int64(0)
			for {
				di, ok := iter.Next()
				if !ok {
					break
				}
				if len(di.Keys) > 0 {
					count += len(di.Keys)
					size += di.Size
					files, ok := geoindex.Extract(di)
					if ok {
						allEXtracts := make([]*scale.OutputFile, 0, len(files))
						for _, ef := range files {
							outName := fmt.Sprintf("%s/extract-file-%s.json", outdir, uuid.NewV4().String())
							WriteJson(outName, ef)
							myOutputFile := &scale.OutputFile{
								Path: outName,
							}
							allEXtracts = append(allEXtracts, myOutputFile)
						}
						outData.Name = "extract_instructions"
						outData.Files = allEXtracts
					}
				}
			}
			log.Println("Processed", humanize.Comma(int64(count)), "items, size:", humanize.Bytes(uint64(size)))
		case geoindex.ExtractFileType:
			file := cr.File.File
			log.Println("Processing", cr.File.File)
			bf := file.AsBucketFile()
			stream := bf.Stream()
			bounds, prj, err := geoindex.DetectType(stream)
			if err != nil {
				WriteStderr(fmt.Sprintf("Error %v", err))
				os.Exit(60)
			}
			ended := time.Now().UTC()
			data := &scale.BoundsResult{Bounds: bounds, Prj: prj, Bucket: bf.Bucket, Key: bf.Key, Region: bf.Region, LastModified: bf.LastModified}
			outName := fmt.Sprintf("%s/bounds_result.json", outdir)
			WriteJson(outName, data)
			outData.Name = "bounds_result"
			outData.File = &scale.OutputFile{Path: outName, GeoMetadata: &scale.GeoMetadata{Started: started.Format(bucket.ISO8601FORMAT), Ended: ended.Format(bucket.ISO8601FORMAT)}}
		default:
			WriteStderr(fmt.Sprintf("Unknown request type %d", cr.RequestType))
			os.Exit(70)
		}
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
			bounds, prj, err1 := geoindex.DetectType(f)
			f.Close()
			if err1 != nil {
				log.Println(err1)
				os.Exit(20)
			}
			log.Println(bounds, prj)
		}
		os.Exit(0)
	}
}
