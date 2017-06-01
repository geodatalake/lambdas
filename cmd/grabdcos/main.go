package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func main() {
	token := flag.Bool("token", false, "Grab DCOS token")
	dir := flag.String("home", "/Users/singram", "Home directory to look in")
	flag.Parse()

	filename := path.Join(*dir, ".dcos", "dcos.toml")
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening", filename, err)
		os.Exit(10)
	}

	r := bufio.NewReader(f)
	for {
		line, err1 := r.ReadString('\n')
		if err1 != nil && err1 != io.EOF {
			f.Close()
			fmt.Println("Error reading", filename, err1)
			os.Exit(20)
		}
		if strings.HasSuffix(line, "\n") {
			line = line[:len(line)-1]
		}
		if strings.Contains(line, "=") {
			vals := strings.Split(line, "=")
			key := strings.TrimSpace(vals[0])
			value := strings.TrimSpace(vals[1])
			value = strings.Trim(value, "\"")
			if *token {
				if key == "dcos_acs_token" {
					fmt.Println(value)
					os.Exit(0)
				}
			} else if key == "dcos_url" {
				fmt.Println(value)
				os.Exit(0)
			}
		}
		if err1 == io.EOF {
			os.Exit(5)
		}
	}
}
