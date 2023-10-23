access-log-parser
=================

[![CI](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/nekrassov01/access-log-parser/graph/badge.svg?token=RIV62CQILM)](https://codecov.io/gh/nekrassov01/access-log-parser)
[![Go Reference](https://pkg.go.dev/badge/github.com/nekrassov01/access-log-parser.svg)](https://pkg.go.dev/github.com/nekrassov01/access-log-parser)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/access-log-parser)](https://goreportcard.com/report/github.com/nekrassov01/access-log-parser)

Simple access log parser utilities written in Go

Usage
-----

```go
func main() {
	// Instantiate Parser with fields and patterns set
	// The default is to return each matched line in NDJSON (newline-separated JSON) format
	p := parser.New(fields, patterns)

	// Parses from a string passed
	log := `dummy string`
	res, err := p.ParseString(log, nil)

	// Read and parse logs from a file
	// All ParseXXX methods can skip lines by specifying the line numbers
	res, err := p.ParseFile("path/to/logfile.log", []int{1, 2})

	// Read and parse logs directly from the gzip file
	res, err := p.ParseGzip("path/to/logfile.log.gz", []int{1, 2})

	// Read logs from a zip file. Default is to read all zip entries, but glob patterns can be applied
	res, err := p.ParseZipEntries("path/to/logfile.log.zip", nil, "*.log")
}
```

Output format
-------------

```go
// Result represents processed data. The Data field contains
// string representations serialized by the designated handler or
// a default handler if none is specified
type Result struct {
	Data     []string `json:"data"`
	Metadata string   `json:"metadata"`
}

// Metadata contains aggregate information about the processed data
// and the lines that did not match
type Metadata struct {
	Total     int           `json:"total"`
	Matched   int           `json:"matched"`
	Unmatched int           `json:"unmatched"`
	Skipped   int           `json:"skipped"`
	Source    string        `json:"source"`
	Errors    []ErrorRecord `json:"errors"`
}

// ErrorRecord represents a record that did not match
type ErrorRecord struct {
	Index  int    `json:"index"`
	Record string `json:"record"`
}
```

Customize
---------

Processing of each matched line and metadata output can be overridden when Parser instantiation.

```go
p := parser.New(fields, patterns, WithLineHandler(customLineHandler), WithMetadataHandler(customMetadataHandler))
```

If you want to pretty-print json, you can wrap the default handler:

```go
func prettyJSON(s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return "", fmt.Errorf("cannot format string as json: %w", err)
	}
	return buf.String(), nil
}

func prettyJSONLineHandler(matches []string, fields []string, index int) (string, error) {
	s, err := parser.DefaultLineHandler(matches, fields, index)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}

func prettyJSONMetadataHandler(m *parser.Metadata) (string, error) {
	s, err := parser.DefaultMetadataHandler(m)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}

p := parser.New(fields, patterns, WithLineHandler(prettyJSONLineHandler), WithMetadataHandler(prettyJSONMetadataHandler))
```

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/access-log-parser/blob/main/LICENSE)
