// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/geodatalake/lambdas/geoindex"
)

func main() {
	flag.Parse()

	args := flag.Args()
	for _, input := range args {
		f, err := os.Open(input)
		if err != nil {
			log.Println("Unable to open", input)
			os.Exit(10)
		}
		bounds, err1 := geoindex.DetectType(f)
		f.Close()
		if err1 != nil {
			log.Println(err1)
			os.Exit(10)
		}
		log.Println(bounds)
		os.Exit(0)
	}
}
