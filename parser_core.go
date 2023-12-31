// Package parser provides utilities for parsing access logs and converting them to structured formats.
package parser

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Parser interface defines the methods for different parsers capable of parsing
// log files in various formats. Each parser can have custom line and metadata
// handlers to process the log lines and metadata respectively.
type Parser interface {
	SetLineHandler(LineHandler)
	SetMetadataHandler(MetadataHandler)
	Parse(input io.Reader, skipLines []int, hasIndex bool) (*Result, error)
	ParseString(input string, skipLines []int, hasIndex bool) (*Result, error)
	ParseFile(input string, skipLines []int, hasIndex bool) (*Result, error)
	ParseGzip(input string, skipLines []int, hasIndex bool) (*Result, error)
	ParseZipEntries(input string, skipLines []int, hasIndex bool, globPattern string) ([]*Result, error)
}

// ErrorRecord stores information about log lines that couldn't be parsed
// according to the provided patterns. This helps in tracking and analyzing
// log lines that do not conform to expected formats.
type ErrorRecord struct {
	Index  int    `json:"index"`  // Index is the index of the log line that didn't match.
	Record string `json:"record"` // Record is the actual log line that was unmatched.
}

// Metadata aggregates information about the parsing process, including
// the total number of lines processed, number of lines matched or skipped,
// and any error records for lines that couldn't be matched.
type Metadata struct {
	Total     int           `json:"total"`     // Total is the total number of log line processed.
	Matched   int           `json:"matched"`   // Matched is the number of log line that successfully matched a pattern.
	Unmatched int           `json:"unmatched"` // Unmatched is the number of log line that did not match any pattern.
	Skipped   int           `json:"skipped"`   // Skipped is the number of log line skipped by the method argument.
	Source    string        `json:"source"`    // Source is the source from which the log was read.
	Errors    []ErrorRecord `json:"errors"`    // Errors is the details of the unmatched record.
}

// Result encapsulates the parsed results of the logs.
type Result struct {
	Data     []string   `json:"data"`     // Data contains the processed log lines, formatted according to the logic defined in LineHandler.
	Metadata string     `json:"metadata"` // Metadata is information about the parsing process, including details of unmatched lines, formatted by the MetadataHandler.
	Labels   [][]string `json:"labels"`   // Labels holds the extracted labels.
	Values   [][]string `json:"values"`   // Values holds the extracted values.
}

// decoder is a function type that decodes a given log line using regular expression patterns.
// It returns slices of labels and values extracted from the line, along with any error encountered.
type decoder func(line string, patterns []*regexp.Regexp) ([]string, []string, error)

// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line index, and returns processed string data.
type LineHandler func(labels []string, values []string, index int, hasIndex bool) (string, error)

// MetadataHandler is a function type for processing and formatting metadata.
type MetadataHandler func(metadata *Metadata) (string, error)

// parse serves as a generic parsing function. It reads from the provided io.Reader,
// applies the given parser function, and handles lines and metadata using the
// specified handlers. It returns a Result object containing parsed data and metadata.
func parse(input io.Reader, skipLines []int, hasIndex bool, decoder decoder, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	data, metadata, labels, values, err := parser(input, skipLines, hasIndex, patterns, decoder, lineHandler)
	if err != nil {
		return nil, err
	}
	return createResult(data, metadata, labels, values, metadataHandler)
}

// parseString provides a convenience method for parsing a string input.
// It wraps the string in a reader and delegates to the generic parse function.
func parseString(input string, skipLines []int, hasIndex bool, decoder decoder, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	return parse(strings.NewReader(input), skipLines, hasIndex, decoder, patterns, lineHandler, metadataHandler)
}

// parseFile handles the parsing of a log file. It opens the file, reads its content,
// and uses the provided parser function to parse the content. It returns a Result
// object containing parsed data and metadata about the file.
func parseFile(input string, skipLines []int, hasIndex bool, decoder decoder, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	data, metadata, labels, values, err := parser(f, skipLines, hasIndex, patterns, decoder, lineHandler)
	if err != nil {
		return nil, err
	}
	metadata.Source = filepath.Base(input)
	return createResult(data, metadata, labels, values, metadataHandler)
}

// parseGzip manages the parsing of gzipped log files. It opens and decompresses the
// gzipped file, then parses its content using the provided parser function. The
// resulting parsed data and metadata are returned in a Result object.
func parseGzip(input string, skipLines []int, hasIndex bool, decoder decoder, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("cannot create gzip reader for %s: %w", input, err)
	}
	defer g.Close()
	data, metadata, labels, values, err := parser(g, skipLines, hasIndex, patterns, decoder, lineHandler)
	if err != nil {
		return nil, err
	}
	metadata.Source = filepath.Base(input)
	return createResult(data, metadata, labels, values, metadataHandler)
}

// parseZipEntries handles parsing of log files within a zip archive. It opens the
// zip file, iterates over its entries, and parses those matching the specified
// glob pattern. Each parsed entry is returned as a Result in a slice.
func parseZipEntries(input string, skipLines []int, hasIndex bool, globPattern string, decoder decoder, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) ([]*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path detected")
	}
	z, err := zip.OpenReader(input)
	if err != nil {
		return nil, fmt.Errorf("cannot open zip file: %w", err)
	}
	defer z.Close()
	var res []*Result
	for _, f := range z.File {
		matched, err := filepath.Match(globPattern, f.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %w", err)
		}
		if !matched {
			continue
		}
		e, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot open zip file entry: %w", err)
		}
		data, metadata, labels, values, err := parser(e, skipLines, hasIndex, patterns, decoder, lineHandler)
		if err != nil {
			return nil, err
		}
		metadata.Source = f.Name
		out, err := createResult(data, metadata, labels, values, metadataHandler)
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}

// parser is a utility function that parses log lines from an io.Reader using the specified decoder and line handler.
// It accumulates parsed data, labels, values, and metadata. Returns a composite data structure containing these elements.
func parser(input io.Reader, skipLines []int, hasIndex bool, patterns []*regexp.Regexp, decoder decoder, handler LineHandler) (data []string, metadata *Metadata, labels [][]string, values [][]string, err error) {
	metadata = &Metadata{}
	i := 1
	m := skipLineMap(skipLines)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if _, ok := m[i]; ok {
			i++
			metadata.Skipped++
			continue
		}
		r := scanner.Text()
		ls, vs, err := decoder(r, patterns)
		if err != nil {
			if err.Error() == "cannot parse input: no pattern provided" {
				return nil, nil, nil, nil, err
			}
			metadata.Errors = append(metadata.Errors, ErrorRecord{Index: i, Record: r})
			metadata.Unmatched++
			i++
			continue
		}
		labels = append(labels, ls)
		values = append(values, vs)
		line, err := handler(ls, vs, i, hasIndex)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		data = append(data, line)
		metadata.Matched++
		i++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("cannot read stream: %w", err)
	}
	metadata.Total = i - 1
	return data, metadata, labels, values, nil
}

// regexDecoder applies regular expression patterns to a given string and
// extracts matching groups. It returns slices of labels and values extracted
// from the string. If no pattern matches, it returns an error.
func regexDecoder(line string, patterns []*regexp.Regexp) ([]string, []string, error) {
	if len(patterns) == 0 {
		return nil, nil, fmt.Errorf("cannot parse input: no pattern provided")
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(line)
		if matches != nil {
			return pattern.SubexpNames()[1:], matches[1:], nil
		}
	}
	return nil, nil, fmt.Errorf("cannot parse line: no matching pattern for line: %s", line)
}

// ltsvDecoder parses a string formatted in Labeled Tab-separated Values (LTSV)
// format. It splits the string into fields based on tabs and then further
// splits each field into labels and values. Returns an error for invalid fields.
func ltsvDecoder(line string, _ []*regexp.Regexp) ([]string, []string, error) {
	fields := strings.Split(line, "\t")
	ls := make([]string, 0, len(fields))
	vs := make([]string, 0, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, ":", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("cannot parse input: invalid field detected: %s", field)
		}
		ls = append(ls, parts[0])
		vs = append(vs, parts[1])
	}
	return ls, vs, nil
}

// createResult aggregates parsed log data and metadata into a Result structure.
// It formats the metadata using the provided MetadataHandler and combines it with the parsed data, labels, and values.
func createResult(data []string, metadata *Metadata, labels [][]string, values [][]string, handler MetadataHandler) (*Result, error) {
	meta, err := handler(metadata)
	if err != nil {
		return nil, err
	}
	return &Result{
		Data:     data,
		Metadata: meta,
		Labels:   labels,
		Values:   values,
	}, nil
}

// skipLineMap generates a map indicating which line numbers should be skipped during parsing.
// It takes a slice of line numbers to skip and returns a map with these line numbers as keys.
func skipLineMap(skipLines []int) map[int]struct{} {
	m := make(map[int]struct{}, len(skipLines))
	for _, skipLine := range skipLines {
		m[skipLine] = struct{}{}
	}
	return m
}
