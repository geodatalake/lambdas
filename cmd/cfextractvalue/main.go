package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type OutputType struct {
	OutputKey   string
	OutputValue string
}

type Stack struct {
	StackId     string
	Outputs     []OutputType
	StackStatus string
}

type Stacks struct {
	Stacks []Stack
}

func main() {
	flag.Parse()

	var key string
	args := flag.Args()
	switch len(args) {
	case 0:
		os.Stderr.WriteString("usage: cfextractvalue key")
		os.Exit(-1)
	case 1:
		key = args[0]
	default:
		os.Stderr.WriteString("usage: cfextractvalue key")
		os.Exit(-1)
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to read input: %v", err))
		os.Exit(10)
	}
	var data Stacks

	if err1 := json.Unmarshal(b, &data); err1 != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse input: %v", err1))
		os.Exit(10)
	}
	if data.Stacks[0].StackStatus == "CREATE_COMPLETE" {
		for _, o := range data.Stacks[0].Outputs {
			if o.OutputKey == key {
				_, value := path.Split(o.OutputValue)
				fmt.Println(value)
				os.Exit(0)
			}
		}
		os.Exit(0)
	} else {
		os.Exit(2)
	}
}
