package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func usage() {
	u := `usage: reformatjson [-sip] inputfile [outputfile]
		version 1.0

		-s Output as a single line
		-p Prefix to use [default" ""]
		-i Indent to use [default: "  "]
		outputfile is optional, if ommitted, stdout is used for output`
	log.Println(u)
}

func openInput(input string) *os.File {
	f, err := os.Open(input)
	if err != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("Unable to open %s: %v", input, err)))
		os.Exit(10)
	}
	return f
}

func openOutput(output string) *os.File {
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("Unable to open %s: %v", output, err)))
			os.Exit(10)
		}
		return f
	} else {
		return os.Stdout
	}
}

func main() {
	singleLine := flag.Bool("s", false, "Output single line")
	prefix := flag.String("p", "", "Prefix to use")
	indent := flag.String("i", "  ", "Indent to use")
	flag.Parse()

	var input, output string
	args := flag.Args()
	switch len(args) {
	case 0:
		usage()
	case 1:
		input = args[0]
		output = ""
	case 2:
		input = args[0]
		output = args[1]
	default:
		usage()
	}
	f := openInput(input)
	var bucket map[string]interface{}
	b, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("Unable to read %s: %v", input, err)))
		os.Exit(10)
	}
	if err2 := json.Unmarshal(b, &bucket); err2 != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("Unable to Unmarshal %s: %v", input, err2)))
		os.Exit(10)
	}
	myOut := openOutput(output)
	if *singleLine {
		b, err3 := json.Marshal(bucket)
		if err3 != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("Unable to Marshal to stdout: %v", err3)))
			os.Exit(10)
		}
		myOut.Write(b)
		os.Exit(0)
	} else {
		b, err3 := json.MarshalIndent(bucket, *prefix, *indent)
		if err3 != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("Unable to Marshal to stdout: %v", err3)))
			os.Exit(10)
		}
		myOut.Write(b)
		os.Exit(0)
	}
}
