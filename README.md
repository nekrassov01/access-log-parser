access-log-parser
=================

Simple access log parser utilities written in Go

Usage
-----

The case of parsing Amazon S3 access logs

```go
package main

import (
	"regexp"
	"strings"

	"github.com/nekrassov01/access-log-parser"
)

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
		`^([!-~]+)`,
		`([!-~]+)`,
		`(\[[ -~]+ [0-9+]+\])`,
		`([!-~]+)`,
		`([!-~]+)`,
		`([!-~]+)`,
		`([!-~]+)`,
		`([!-~]+)`,
		`"([ -~]+)"`,
		`(\d{1,3})`,
		`([!-~]+)`,
		`([\d\-.]+)`,
		`([\d\-.]+)`,
		`([\d\-.]+)`,
		`([\d\-.]+)`,
		`"([ -~]+)"`,
		`"([ -~]+)"`,
		`([!-~]+)`,
	}

	patternV2 = append(
		patternV1,
		[]string{
			`([!-~]+)`,
			`([!-~]+)`,
			`([!-~]+)`,
			`([!-~]+)`,
			`([!-~]+)`,
		}...,
	)

	patternV3 = append(
		patternV2,
		[]string{
			`([!-~]+)`,
		}...,
	)

	patternV4 = append(
	patternV3,
		[]string{
			`([!-~]+)$`,
		}...,
	)

	s3PatternV5 = append(
		patternV4,
		[]string{
			`([!-~]+)$`,
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

// Function to instantiate a parser by setting fields and patterns
func NewS3LogsJSONParser() *parser.Parser {
	return parser.New(fields, patterns, nil, nil)
}

func main() {
	p := NewS3LogsJSONParser()

    // Read and parse logs from a file
	res, err := p.ParseFile("path/to/logfile.log", []int{1, 2})

    // Read and parse logs forom gzip file
	res, err := p.ParseGzip("path/to/logfile.log.gz", []int{1, 2})

    // Read logs from a ZIP file. Defaults to all ZIP entries, but the glob pattern can be applied
	res, err := p.ParseZipEntries("path/to/logfile.log.zip", nil, "*.log")
}
```

Customize
---------

Processing for each matched line and outputting metadata can be overridden during Parser initialization.

```go
func NewS3LogsJSONParser() *parser.Parser {
	return parser.New(fields, patterns, customLineHandler, customMetadataHandler)
}
```

If you want to pretty-print json, just wrap the default handler:

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
```

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/access-log-parser/blob/main/LICENSE)
