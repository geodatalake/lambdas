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

	"github.com/dustin/go-humanize"
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

func produceJobTypeBucket() []byte {
	data := doc().
		AddKV("name", "open-bucket").
		AddKV("version", "1.0.0").
		AddKV("title", "Open Bucket Geo").
		AddKV("description", "Opens a S3 Bucket by directory").
		AddKV("category", "testing").
		AddKV("author_name", "Steve_Ingram").
		AddKV("author_url", "http://www.example.com").
		AddKV("is_operational", true).
		AddKV("icon_code", "f2b7").
		AddKV("docker_privileged", false).
		AddKV("docker_image", "openwhere/scale-detector:dev").
		AddKV("priority", 230).
		AddKV("max_tries", 3).
		AddKV("cpus_required", 1.0).
		AddKV("mem_required", 8192.0).
		AddKV("disk_out_const_required", 0.0).
		AddKV("disk_out_mult_required", 0.0).
		Append("interface", doc().
			AddKV("version", "1.1").
			AddKV("command", "/opt/detect/detector").
			AddKV("command_arguments", "${open_bucket} ${job_output_dir}").
			AddKV("shared_resources", []map[string]interface{}{}).
			AppendArray("output_data", array().
				Add(doc().
					AddKV("media_type", "application/json").
					AddKV("required", true).
					AddKV("type", "files").
					AddKV("name", "extract_instructions"))).
			AppendArray("input_data", array().
				Add(doc().
					AppendArray("media_types", array().
						Add("application/json")).
					AddKV("required", true).
					AddKV("partial", false).
					AddKV("type", "file").
					AddKV("name", "open_bucket")))).
		Append("error_mapping", doc().
			AddKV("version", "1.0").
			Append("exit_codes", doc().
				AddKV("10", "bad_num_input").
				AddKV("20", "open_input").
				AddKV("30", "read_input").
				AddKV("40", "marshal_failure").
				AddKV("50", "bad_s3_read").
				AddKV("70", "bad_cluster_request")))
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
	scale.CreateScaleError(url, token, scale.ErrorDoc("bad_cluster_request", "Invalid Cluster Request", "Unknown cluster request", existing))
}

func register(url, token string, data []byte) {
	client := http.Client{}
	urlStr := fmt.Sprintf("%s/service/scale/api/v5/job-types/", url)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
	req.Header.Add("Authorization", fmt.Sprintf("token=%s", token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		scale.WriteStderr(fmt.Sprintf("Error registering job type: %v", err))
		os.Exit(-1)
	}
	if resp.StatusCode != 201 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			scale.WriteStderr(err.Error())
			os.Exit(-1)
		}
		resp.Body.Close()
		fmt.Println(resp.Status, string(b))
	} else {
		resp.Body.Close()
		fmt.Println("Create Job Type Response:", resp.Status)
	}
}

func registerJobTypes(url, token string) {
	// Errors have to registered prior to job type ref'ing those errors
	createErrors(url, token)
	register(url, token, produceJobTypeBucket())
}

//  Errors:
// 10 Bad number of input arguments
// 20 Unable to open input
// 30 Unable to read input
// 40 Unable to marshal cluster request
// 50 Unable to read S3 bucket
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
		fmt.Println(string(produceJobTypeBucket()))
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
		outData := new(scale.OutputData)
		switch cr.RequestType {
		case geoindex.ScanBucket:
			root, err2 := bucket.ListBucketStructure(cr.Bucket.Region, cr.Bucket.Bucket)
			if err2 != nil {
				scale.WriteStderr(err2.Error())
				os.Exit(50)
			}
			iter := root.Iterate()
			count := 0
			size := int64(0)
			allEXtracts := make([]*scale.OutputFile, 0, 128)
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
						for _, ef := range files {
							outName := fmt.Sprintf("%s/extract-file-%s.json", outdir, uuid.NewV4().String())
							scale.WriteJson(outName, ef)
							myOutputFile := &scale.OutputFile{
								Path: outName,
							}
							allEXtracts = append(allEXtracts, myOutputFile)
						}
					}
				}
			}
			log.Println("Processed", humanize.Comma(int64(count)), "items, size:", humanize.Bytes(uint64(size)))
			outData.Name = "extract_instructions"
			outData.Files = allEXtracts
		default:
			scale.WriteStderr(fmt.Sprintf("Unknown request type %d", cr.RequestType))
			os.Exit(70)
		}
		manifest := scale.FormatManifest([]*scale.OutputData{outData}, nil)
		scale.WriteJson(fmt.Sprintf("%s/results_manifest.json", outdir), manifest)
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
			bounds, prj, typ, err1 := geoindex.DetectType(f)
			f.Close()
			if err1 != nil {
				log.Println(err1)
				os.Exit(20)
			}
			log.Println(bounds, prj, typ)
		}
		os.Exit(0)
	}
}
