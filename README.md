access-log-parser
=================

[![CI](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/nekrassov01/access-log-parser/graph/badge.svg?token=RIV62CQILM)](https://codecov.io/gh/nekrassov01/access-log-parser)
[![Go Reference](https://pkg.go.dev/badge/github.com/nekrassov01/access-log-parser.svg)](https://pkg.go.dev/github.com/nekrassov01/access-log-parser)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/access-log-parser)](https://goreportcard.com/report/github.com/nekrassov01/access-log-parser)

Simple access log parser utilities written in Go

Usage
-----

`RegexParser` allows multiple regular expression patterns to be specified and matched in sequence.

```go
func main() {
	// Instantiate Parser
	p := parser.NewRegexParser()

	// Multiple patterns can be set at the same time and matched in order
	// Each pattern must have at least one named capture group
	if err := p.AddPatterns(patterns); err != nil {
		log.Fatal(err)
	}

	// Parses from a string passed
	log := `dummy string`
	res, err := p.ParseString(log, nil, false)

	// Read and parse logs from a file
	// All ParseXXX methods can skip lines by specifying the line numbers
	res, err := p.ParseFile("path/to/logfile.log", []int{1, 2}, false)

	// Read and parse logs directly from the gzip file
	res, err := p.ParseGzip("path/to/logfile.log.gz", []int{1, 2}, false)

	// Read logs from a zip file. Default is to read all zip entries, but glob patterns can be applied
	res, err := p.ParseZipEntries("path/to/logfile.log.zip", nil, false, "*.log")

	// If you want to add an Index (Line number) at the beginning of a line, set the third argument to true
	res, err := p.ParseFile("path/to/logfile.log", nil, true)

	// By passing a single line, labels can be extracted.
	// If there is a line break, the labels are obtained for the first element of the division.
    // The second argument can specify whether to have line numbers as items.
	labels, err := p.Label(log, true)
}
```

`LTSVParser` can parse logs in LTSV format and convert them to other structured formats.

```go
func main() {
	// Instantiate Parser
	p := parser.NewLTSVParser()

	// Parses from a LTSV line passed
	log := `label1:value1	label2:value2	label3:value3`
	res, err := p.ParseString(log, nil, false)

	// Can use the same ParseXXX method as RegexParser
}
```

Output format
-------------

```go
// Result represents processed data. Data and Metadata are set
// to the data serialized by the respective handlers.
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

The processing of each matched row and the output of metadata can be overridden by setter methods.

```go
p = parser.NewRegexParser()
p.SetLineHandler(yourCustomLineHandler),
p.SetMetadataHandler(yourCustomMetadataHandler),
```

The following handlers are preset:

- JSON (default): `JSONLineHandler`, `JSONMetadataHandler`
- Pretty JSON: `PrettyJSONLineHandler`, `PrettyJSONMetadataHandler`
- key=value pair: `KeyValuePairLineHandler`, `KeyValuePairMetadataHandler`
- LTSV: `LTSVLineHandler`, `LTSVMetadataHandler`
- TSV: `TSVLineHandler`, `TSVMetadataHandler`

Preset Constructors
-------------------

Functions are provided by default to instantiate the following parsers:

- Apache common/combined log format: `NewApacheCLFRegexParser()`
- Apache common/combined log format with virtual host: `NewApacheCLFWithVHostRegexParser()`
- Amazon S3 access log format: `NewS3RegexParser()`
- Amazon CloudFront access log format: `NewCFRegexParser()`
- AWS Application Load Balancer access log format: `NewALBRegexParser()`
- AWS Network Load Balancer access log format: `NewNLBRegexParser()`
- AWS Classic Load Balancer access log format: `NewCLBRegexParser()`

Sample
------

[__alpen__](https://github.com/nekrassov01/alpen/blob/main/app.go#L353-L395) is an application for parsing and encoding various access logs.

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/access-log-parser/blob/main/LICENSE)
