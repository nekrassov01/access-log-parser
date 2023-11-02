package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nekrassov01/access-log-parser"
)

var (
	patterns = []*regexp.Regexp{
		regexp.MustCompile("(?P<field1>[!-~]+) (?P<field2>[!-~]+) (?P<field3>[!-~]+)"),
	}
	sample = `aaa bbb ccc
xxx yyy zzz
111 222 333`
)

func main() {
	// default
	p := logparser.NewParser()
	if err := p.AddPatterns(patterns); err != nil {
		log.Fatal(err)
	}
	out, err := p.Parse(strings.NewReader(sample), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)

	/*
		{"index":1,"field1":"aaa","field2":"bbb","field3":"ccc"}
		{"index":2,"field1":"xxx","field2":"yyy","field3":"zzz"}
		{"index":3,"field1":"111","field2":"222","field3":"333"}
		{"total":3,"matched":3,"unmatched":0,"skipped":0,"source":"","errors":null}
	*/

	// customize
	p = logparser.NewParser(
		logparser.WithLineHandler(prettyJSONLineHandler),
		logparser.WithMetadataHandler(prettyJSONMetadataHandler),
	)
	if err := p.AddPatterns(patterns); err != nil {
		log.Fatal(err)
	}
	out, err = p.Parse(strings.NewReader(sample), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)

	/*
		{
			"index": 1,
			"field1": "aaa",
			"field2": "bbb",
			"field3": "ccc"
		}
		{
			"index": 2,
			"field1": "xxx",
			"field2": "yyy",
			"field3": "zzz"
		}
		{
			"index": 3,
			"field1": "111",
			"field2": "222",
			"field3": "333"
		}
		{
			"total": 3,
			"matched": 3,
			"unmatched": 0,
			"skipped": 0,
			"source": "",
			"errors": null
		}
	*/
}

func prettyJSON(s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return "", fmt.Errorf("cannot format string as json: %w", err)
	}
	return buf.String(), nil
}

func prettyJSONLineHandler(matches []string, fields []string, index int) (string, error) {
	s, err := logparser.DefaultLineHandler(matches, fields, index)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}

func prettyJSONMetadataHandler(m *logparser.Metadata) (string, error) {
	s, err := logparser.DefaultMetadataHandler(m)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}
