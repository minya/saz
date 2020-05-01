package main

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

var reFileName *regexp.Regexp

func init() {
	reFileName, _ = regexp.Compile("(\\d+)_(\\w)")
}

type Session struct {
	XMLName xml.Name      `xml:"Session"`
	Timers  SessionTimers `xml:"SessionTimers"`
	Flags   SessionFlags  `xml:"SessionFlags"`
}

type SessionTimers struct {
	XMLName             xml.Name `xml:"SessionTimers"`
	ClientConnected     string   `xml:"ClientConnected,attr"`
	ClientBeginRequest  string   `xml:"ClientBeginRequest,attr"`
	GotRequestHeaders   string   `xml:"GotRequestHeaders,attr"`
	ClientDoneRequest   string   `xml:"ClientDoneRequest,attr"`
	GatewayTime         string   `xml:"GatewayTime,attr"`
	DNSTime             string   `xml:"DNSTime,attr"`
	TCPConnectTime      string   `xml:"TCPConnectTime,attr"`
	HTTPSHandshakeTime  string   `xml:"HTTPSHandshakeTime,attr"`
	ServerConnected     string   `xml:"ServerConnected,attr"`
	FiddlerBeginRequest string   `xml:"FiddlerBeginRequest,attr"`
	ServerGotRequest    string   `xml:"ServerGotRequest,attr"`
	ServerBeginResponse string   `xml:"ServerBeginResponse,attr"`
	GotResponseHeaders  string   `xml:"GotResponseHeaders,attr"`
	ServerDoneResponse  string   `xml:"ServerDoneResponse,attr"`
	ClientBeginResponse string   `xml:"ClientBeginResponse,attr"`
	ClientDoneResponse  string   `xml:"ClientDoneResponse,attr"`
}

type SessionFlags struct {
	XMLName xml.Name      `xml:"SessionFlags"`
	Flags   []SessionFlag `xml:"SessionFlag"`
}

type SessionFlag struct {
	XMLName xml.Name `xml:"SessionFlag"`
	Name    string   `xml:"N,attr"`
	Value   string   `xml:"V,attr"`
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
	var session Session
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

		if t == "m" {
			read, err := f.Open()
			if nil != err {
				fmt.Printf("%v\n", err)
				os.Exit(-1)
			}
			defer read.Close()

			bytes, _ := ioutil.ReadAll(read)
			xml.Unmarshal(bytes, &session)
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

			printResult(num, request, response, session)
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

func printResult(num string, request *http.Request, response *http.Response, session Session) {
	clientBeginRequest, err := time.Parse(time.RFC3339, session.Timers.ClientBeginRequest)
	if nil != err {
		fmt.Println("Error while parsing clientConnected:", err)
		os.Exit(-1)
	}
	clientDoneResponse, err := time.Parse(time.RFC3339, session.Timers.ClientDoneResponse)
	if nil != err {
		fmt.Println("Error while parsing clientDoneResponse:", err)
		os.Exit(-1)
	}
	var process string
	for _, flag := range session.Flags.Flags {
		if flag.Name == "x-processinfo" {
			process = flag.Value
			break
		}
	}
	fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\t%v\n", num, request.Method, response.StatusCode, request.URL.String(),
		clientBeginRequest.Format("15:04:05.000"), clientDoneResponse.Format("15:04:05.000"), process)
}
