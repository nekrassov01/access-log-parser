package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nekrassov01/access-log-parser"
)

var (
	fields   = []string{"field1", "field2", "field3"}
	patterns = []*regexp.Regexp{regexp.MustCompile("([!-~]+) ([!-~]+) ([!-~]+)")}
	sample   = `aaa bbb ccc
xxx yyy zzz
111 222 333`
)

func main() {
	p := parser.New(fields, patterns)
	out, err := p.Parse(strings.NewReader(sample), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
}

/*
{"index":1,"field1":"aaa","field2":"bbb","field3":"ccc"}
{"index":2,"field1":"xxx","field2":"yyy","field3":"zzz"}
{"index":3,"field1":"111","field2":"222","field3":"333"}
{"total":3,"matched":3,"unmatched":0,"skipped":0,"source":"","errors":null}
*/
