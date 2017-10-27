package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

var reFileName *regexp.Regexp

func init() {
	reFileName, _ = regexp.Compile("(\\d+)_(\\w)")
}

func main() {
	flag.Parse()
	r, err := zip.OpenReader(flag.Arg(0))
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(-1)
	}
	defer r.Close()
	var request *http.Request
	var response *http.Response
	for _, f := range r.File {
		match, num, t := parseFileName(f.Name)
		if match == false {
			continue
		}
		if t == "c" {
			read, err := f.Open()
			if nil != err {
				fmt.Printf("%v\n", err)
				os.Exit(-1)
			}
			defer read.Close()

			reqReader := bufio.NewReader(read)
			request, _ = http.ReadRequest(reqReader)
		}

		if t == "s" {
			read, err := f.Open()
			if nil != err {
				fmt.Printf("%v\n", err)
				os.Exit(-1)
			}
			defer read.Close()

			respReader := bufio.NewReader(read)
			response, _ = http.ReadResponse(respReader, request)

			printResult(num, request, response)
		}
	}
}

func parseFileName(name string) (bool, string, string) {
	match := reFileName.FindAllStringSubmatch(name, -1)
	if len(match) == 0 {
		return false, "", ""
	}
	return true, match[0][1], match[0][2]
}

func printResult(num string, request *http.Request, response *http.Response) {
	fmt.Printf("%v\t%v\t%v\t%v\n", num, request.Method, response.StatusCode, request.URL.String())
}
