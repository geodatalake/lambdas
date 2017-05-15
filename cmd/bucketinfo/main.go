package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/geodatalake/lambdas/bucket"
)

func main() {
	region := flag.String("r", "us-west-2", "AWS Region")
	countsOnly := flag.Bool("co", false, "Only counts")
	flag.Parse()

	overallSize := int64(0)
	overallItems := int64(0)
	bucketUris := flag.Args()
	for _, uri := range bucketUris {
		log.Println(uri)
		dirs, err := bucket.ListBucketStructure(*region, uri)
		output := make([]string, 0, 128)
		totalItems := 0
		totalSize := int64(0)
		ch := make(chan *bucket.DirInfo, 1)
		go func(c chan *bucket.DirInfo) {
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
					break
				}
			}
		}(ch)
		dirs.Print(0, ch)
		log.Println("Closing channel")
		close(ch)
		if *countsOnly {
			log.Println(strings.Join(output, "\n"))
		}
		log.Println("Total Items Found:", totalItems)
		overallItems += int64(totalItems)
		log.Println("Total Size of Items:", humanize.Bytes(uint64(totalSize)))
		overallSize += totalSize
		if err != nil {
			log.Println(err)
			os.Exit(10)
		}
	}
	log.Println("Overall Items Found:", humanize.Comma(overallItems))
	log.Println("Total Size of Overall Items:", humanize.Bytes(uint64(overallSize)))
}
