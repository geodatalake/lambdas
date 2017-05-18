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

	"github.com/dustin/go-humanize"
	"github.com/geodatalake/lambdas/bucket"
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

func main() {
	dev := flag.Bool("dev", false, "Development flag, interpret input as image file")
	flag.Parse()
	if !*dev {
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
			os.Exit(10)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			WriteStderr(err.Error())
			os.Exit(10)
		}
		f.Close()
		var cr geoindex.ClusterRequest
		if errJson := json.Unmarshal(b, &cr); errJson != nil {
			WriteStderr(errJson.Error())
			os.Exit(10)
		}
		outData := new(scale.OutputData)
		switch cr.RequestType {
		case geoindex.ScanBucket:
			root, err2 := bucket.ListBucketStructure(cr.Bucket.Region, cr.Bucket.Bucket)
			if err2 != nil {
				WriteStderr(err2.Error())
				os.Exit(10)
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
			bf := file.AsBucketFile()
			stream := bf.Stream()
			bounds, prj, err := geoindex.DetectType(stream)
			if err != nil {
				WriteStderr(fmt.Sprintf("Error %v", err))
				os.Exit(10)
			}
			data := &scale.BoundsResult{Bounds: bounds, Prj: prj}
			outName := fmt.Sprintf("%s/bounds_result.json", outdir)
			WriteJson(outName, data)
			outData.Name = "bounds_result"
			outData.File = &scale.OutputFile{Path: outName}
		default:
			WriteStderr(fmt.Sprintf("Unknown request type %d", cr.RequestType))
			os.Exit(50)
		}
		manifest := scale.FormatManifest([]*scale.OutputData{outData}, nil)
		WriteJson(fmt.Sprintf("%s/results_manifest.json", outdir), manifest)
		os.Exit(0)
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
