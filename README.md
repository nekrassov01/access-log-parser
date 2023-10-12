access-log-parser
=================

[![CI](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/nekrassov01/access-log-parser/graph/badge.svg?token=RIV62CQILM)](https://codecov.io/gh/nekrassov01/access-log-parser)
[![Go Reference](https://pkg.go.dev/badge/github.com/nekrassov01/access-log-parser.svg)](https://pkg.go.dev/github.com/nekrassov01/access-log-parser)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/access-log-parser)](https://goreportcard.com/report/github.com/nekrassov01/access-log-parser)

Simple access log parser utilities written in Go

Usage
-----

For example, to parse Amazon S3 access logs:

```go
const sep = " "

var (
	fields = []string{
		"bucket_owner",
		"bucket",
		"time",
		"remote_ip",
		"requester",
		"request_id",
		"operation",
		"key",
		"request_uri",
		"http_status",
		"error_code",
		"bytes_sent",
		"object_size",
		"total_time",
		"turn_around_time",
		"referer",
		"user_agent",
		"version_id",
		"host_id",
		"signature_version",
		"cipher_suite",
		"authentication_type",
		"host_header",
		"tls_version",
		"access_point_arn",
		"acl_required"
	}

	patternV1 = []string{
		`^([!-~]+)`,            // bucket_owner
		`([!-~]+)`,             // bucket
		`(\[[ -~]+ [0-9+]+\])`, // time
		`([!-~]+)`,             // remote_ip
		`([!-~]+)`,             // requester
		`([!-~]+)`,             // request_id
		`([!-~]+)`,             // operation
		`([!-~]+)`,             // key
		`"([ -~]+)"`,           // request_uri
		`(\d{1,3})`,            // http_status
		`([!-~]+)`,             // error_code
		`([\d\-.]+)`,           // bytes_sent
		`([\d\-.]+)`,           // object_size
		`([\d\-.]+)`,           // total_time
		`([\d\-.]+)`,           //turn_around_time
		`"([ -~]+)"`,           // referer
		`"([ -~]+)"`,           // user_agent
		`([!-~]+)`,             // version_id
	}

	patternV2 = append(
		patternV1, []string{
			`([!-~]+)`, // host_id
			`([!-~]+)`, // signature_version
			`([!-~]+)`, // cipher_suite
			`([!-~]+)`, // authentication_type
			`([!-~]+)`, // host_header
		}...,
	)

	patternV3 = append(
		patternV2, []string{
			`([!-~]+)`, // tls_version
		}...,
	)

	patternV4 = append(
		patternV3, []string{
			`([!-~]+)$`, // access_point_arn
		}...,
	)
	s3PatternV5 = append(
		patternV4, []string{
			`([!-~]+)$`, // acl_required
		}...,
	)

	patterns = []*regexp.Regexp{
		regexp.MustCompile(strings.Join(patternV5, sep)),
		regexp.MustCompile(strings.Join(patternV4, sep)),
		regexp.MustCompile(strings.Join(patternV3, sep)),
		regexp.MustCompile(strings.Join(patternV2, sep)),
		regexp.MustCompile(strings.Join(patternV1, sep)),
	}
)

func main() {
	// Instantiate Parser with fields and patterns set
	// The default is to return each matched line in NDJSON (newline-separated JSON) format
	p := parser.New(fields, patterns, nil, nil)

	// Parses from a string passed
	log := `dummy string`
	res, err := p.ParseString(log, nil)

	// Read and parse logs from a file
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
type Result struct {
	Data     []string `json:"data"`
	Metadata string   `json:"metadata"`
}

type Metadata struct {
    Total     int           `json:"total"`
	Matched   int           `json:"matched"`
	Unmatched int           `json:"unmatched"`
	Skipped   int           `json:"skipped"`
	Source    string        `json:"source"`
	Errors    []ErrorRecord `json:"errors"`
}

type ErrorRecord struct {
		Index  int    `json:"index"`
		Record string `json:"record"`
	}
```

Customize
---------

Processing of each matched line and metadata output can be overridden when Parser instantiation.

```go
p := parser.New(fields, patterns, customLineHandler, customMetadataHandler)
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

p := parser.New(fields, patterns, prettyJSONLineHandler, prettyJSONMetadataHandler)
```

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/access-log-parser/blob/main/LICENSE)
