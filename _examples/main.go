package main

import (
	"context"
	"fmt"
	"log"
	"os"

	parser "github.com/nekrassov01/access-log-parser"
)

func main() {
	var p parser.Parser
	var r *parser.Result
	var err error

	p = parser.NewS3RegexParser(os.Stdout)
	p.SetLineHandler(parser.JSONLineHandler)

	/*
		Example for realtime streaming processing

		Args:
		ctx            : context for signal notify
		reader         : input reader
		keywords       : filter lines by keywords
		labels         : select columns by field name
		hasPrefix      : enable display of prefixes such as [ PROCESSED | UNMATCHED ]
		disableUnmatch : disable display of unmatch log line

		$ tail -f testdata/sample_s3_contains_unmatch.log | go run _examples/main.go stream

		Exit with CTRL+C
	*/
	if len(os.Args) > 1 && os.Args[1] == "stream" {
		fmt.Println("\033[1;32mParse\033[0m")
		fmt.Println("\033[1;32m-----\033[0m")
		fmt.Println()
		r, err = p.Parse(context.Background(), os.Stdin, nil, nil, true, false)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(r)
		return
	}

	/*
		Example of parsing from a file path (Same signature for `ParseString` and `ParseGzip`)

		Args:
		filePath      : file path
		keywords      : filter lines by keywords
		labels        : select columns by field name
		skipLines     : skip lines by line number (not index)
		hasLineNumber : enable line number display
	*/
	fmt.Println("\033[1;32mParseFile\033[0m")
	fmt.Println("\033[1;32m---------\033[0m")
	fmt.Println()
	r, err = p.ParseFile(
		"testdata/sample_s3_contains_unmatch.log",
		[]string{"REST.GET.VERSIONING"},
		[]string{"bucket_owner", "bucket", "method", "request_uri", "protocol"},
		[]int{1},
		true,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)

	/*
		Example of parsing from a zip path

		Args:
		filePath      : file path
		globPattern   : glob pattern to filter zip entries
		keywords      : filter lines by keywords
		labels        : select columns by field name
		skipLines     : skip lines by line number (not index)
		hasLineNumber : enable line number display
	*/
	fmt.Println("\033[1;32mParseZipEntries\033[0m")
	fmt.Println("\033[1;32m---------------\033[0m")
	fmt.Println()
	p.SetLineHandler(parser.TSVLineHandler)
	r, err = p.ParseZipEntries("testdata/sample_s3.zip", "*.log", nil, nil, nil, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)

	/*
		Example of parsing from a string
	*/
	fmt.Println("\033[1;32mParseString\033[0m")
	fmt.Println("\033[1;32m-----------\033[0m")
	fmt.Println()
	p = parser.NewLTSVParser(os.Stdout)
	p.SetLineHandler(parser.PrettyJSONLineHandler)
	r, err = p.ParseString(
		`remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)`,
		nil, nil, nil, true,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}
