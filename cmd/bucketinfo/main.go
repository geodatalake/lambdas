package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/geodatalake/lambdas/bucket"
)

func main() {
	region := flag.String("r", "us-west-2", "AWS Region")
	countsOnly := flag.Bool("co", false, "Only counts")
	prefix := flag.String("prefix", "", "Prefix to use")
	versions := flag.Bool("v", false, "Include versions")
	before := flag.String("before", "", "Filter only lastModified before date yyy-MM-dd")
	flag.Parse()

	overallSize := int64(0)
	overallItems := int64(0)
	bucketUris := flag.Args()
	for _, uri := range bucketUris {
		log.Println(uri)
		svc := s3.New(session.New(), &aws.Config{Region: region})
		if *prefix != "" {
			files, err := bucket.ReadBucketDir(*region, uri, *prefix, svc)
			if err != nil {
				log.Println(err)
				os.Exit(10)
			}
			log.Println("Found", len(files), "files")
			for _, f := range files {
				fmt.Printf("%+v\n", *f)
			}
			os.Exit(0)
		}
		filterBefore := time.Now()
		if *before != "" {
			if tm, err := time.Parse("2006-01-02", *before); err == nil {
				filterBefore = tm
			} else {
				log.Println("Failed to parse", *before, err)
			}
		}
		dirs, err := bucket.ListBucketStructure(*region, uri, svc, *versions, filterBefore)
		output := make([]string, 0, 128)
		totalItems := 0
		totalSize := int64(0)
		var wg sync.WaitGroup
		ch := make(chan *bucket.DirInfo, 1024)
		wg.Add(1)
		go func(c chan *bucket.DirInfo, waiter *sync.WaitGroup) {
			for {
				info, good := <-c
				if good {
					s := strings.Repeat("  ", info.Level) + info.Name
					if info.Keys > 0 {
						totalItems += info.Keys
						totalSize += info.Size
						plural := "item"
						if info.Keys > 1 {
							plural = "items"
						}
						s = fmt.Sprintf("%s %d %s", s, info.Keys, plural)
					}
					if !*countsOnly {
						output = append(output, s)
					}
				} else {
					log.Println("Channel is closed")
					waiter.Done()
					break
				}
			}
		}(ch, &wg)
		dirs.Print(0, ch)
		log.Println("Closing channel")
		close(ch)
		wg.Wait()
		if !*countsOnly {
			log.Println(strings.Join(output, "\n"))
		}
		log.Println("Total Items Found:", totalItems)
		overallItems += int64(totalItems)
		log.Println("Total Size of Items:", humanize.Bytes(uint64(totalSize)), humanize.Comma(totalSize))
		overallSize += totalSize
		if err != nil {
			log.Println(err)
			os.Exit(10)
		}
	}
	log.Println("Overall Items Found:", humanize.Comma(overallItems))
	log.Println("Total Size of Overall Items:", humanize.Bytes(uint64(overallSize)), humanize.Comma(overallSize))
}
