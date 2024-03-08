access-log-parser
=================

[![CI](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/nekrassov01/access-log-parser/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/nekrassov01/access-log-parser/graph/badge.svg?token=RIV62CQILM)](https://codecov.io/gh/nekrassov01/access-log-parser)
[![Go Reference](https://pkg.go.dev/badge/github.com/nekrassov01/access-log-parser.svg)](https://pkg.go.dev/github.com/nekrassov01/access-log-parser)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/access-log-parser)](https://goreportcard.com/report/github.com/nekrassov01/access-log-parser)

Simple access log parser utilities written in Go

Features
--------

- Flexible serialization of log lines
- Streaming processing support
- Line filtering by filter expressions like `size < 100` `method == GET` `remote_host =~ ^192.168.`
- Display column selection by field name
- Line skipping by line number
- Customization by handler functions
- Various preset constructors for well-known log formats
- LTSV format support

Usage
-----

[Example](https://github.com/nekrassov01/access-log-parser/blob/main/examples_test.go)

Output format
-------------

Parsed log lines are sequentially output to Writer. After the parsing is finished, the total result is output.

```go
// Result encapsulates the outcomes of parsing operations, detailing matched, unmatched, excluded,
// and skipped line counts, along with processing time and source information.
type Result struct {
	Total       int           `json:"total"`                // Total number of processed lines.
	Matched     int           `json:"matched"`              // Count of lines that matched the patterns.
	Unmatched   int           `json:"unmatched"`            // Count of lines that did not match any patterns.
	Excluded    int           `json:"excluded"`             // Count of lines excluded based on keyword search.
	Skipped     int           `json:"skipped"`              // Count of lines skipped explicitly.
	ElapsedTime time.Duration `json:"elapsedTime"`          // Processing time for the log data.
	Source      string        `json:"source"`               // Source of the log data.
	ZipEntries  []string      `json:"zipEntries,omitempty"` // List of processed zip entries, if applicable.
	Errors      []Errors      `json:"errors"`               // Collection of errors encountered during parsing.
	inputType   inputType     `json:"-"`                    // Type of input being processed.
}

// Errors stores information about log lines that couldn't be parsed
// according to the provided patterns. This helps in tracking and analyzing
// log lines that do not conform to expected formats.
type Errors struct {
	Entry      string `json:"entry,omitempty"` // Optional entry name if the log came from a zip file.
	LineNumber int    `json:"lineNumber"`      // Line number of the problematic log entry.
	Line       string `json:"line"`            // Content of the problematic log line.
}
```

The struct `Result` implements `fmt.Stringer` as follows:

```text
/* SUMMARY */

+-------+---------+-----------+----------+---------+-------------+--------------------------------+
| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime | Source                         |
+-------+---------+-----------+----------+---------+-------------+--------------------------------+
|     5 |       4 |         1 |        0 |       0 | 1.16375ms   | sample_s3_contains_unmatch.log |
+-------+---------+-----------+----------+---------+-------------+--------------------------------+

Total     : Total number of log line processed
Matched   : Number of log line that successfully matched pattern
Unmatched : Number of log line that did not match any pattern
Excluded  : Number of log line that did not extract by filter expressions
Skipped   : Number of log line that skipped by line number (disabled in stream mode)

/* UNMATCH LINES */

+------------+------------------------------------------------------------------------------------------------------+
| LineNumber | Line                                                                                                 |
+------------+------------------------------------------------------------------------------------------------------+
|          4 | d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:5 |
|            | 4:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXA |
|            | MPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Cons |
|            | ole/0.4"                                                                                             |
+------------+------------------------------------------------------------------------------------------------------+

LineNumber : Line number of the log that did not match any pattern
Line       : Raw log line that did not match any pattern
```

Customize
---------

The processing of each matched row can be overridden by a setter method if a `LineHandler` type function is implemented.

```go
p = parser.NewRegexParser(os.Stdout)
p.SetLineHandler(yourCustomLineHandler),
```

Set your function with the following signature:

```go
// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line index, and returns processed string data.
type LineHandler func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error)
```

> [!NOTE]
>The reason we did not use maps is that the measured results were almost identical when the overhead of setting the order keep is taken into account. (However, we did not take a very rigorous benchmark.)

The following handlers are preset:

- JSON (default): `JSONLineHandler`
- Pretty JSON: `PrettyJSONLineHandler`
- key=value pair: `KeyValuePairLineHandler`
- LTSV: `LTSVLineHandler`
- TSV: `TSVLineHandler`

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

[__alpen__](https://github.com/nekrassov01/alpen) is an application for parsing and encoding various access logs.

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/access-log-parser/blob/main/LICENSE)
