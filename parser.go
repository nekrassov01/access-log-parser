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

// Version of access-log-parser.
const Version = "0.0.6"

type Parser interface {
	SetLineHandler(LineHandler)
	SetMetadataHandler(MetadataHandler)
	Parse(input io.Reader, skipLines []int) (*Result, error)
	ParseString(input string, skipLines []int) (*Result, error)
	ParseFile(input string, skipLines []int) (*Result, error)
	ParseGzip(input string, skipLines []int) (*Result, error)
	ParseZipEntries(input string, skipLines []int, globPattern string) ([]*Result, error)
}

// ErrorRecord represents a record that didn't match any of the provided patterns.
// Such records are stored in this structure to prevent data loss and for further inspection.
type ErrorRecord struct {
	Index  int    `json:"index"`  // Index is the index of the log line that didn't match.
	Record string `json:"record"` // Record is the actual log line that was unmatched.
}

// Metadata represents metadata about the parsed records.
type Metadata struct {
	Total     int           `json:"total"`     // Total is the total number of log line processed.
	Matched   int           `json:"matched"`   // Matched is the number of log line that successfully matched a pattern.
	Unmatched int           `json:"unmatched"` // Unmatched is the number of log line that did not match any pattern.
	Skipped   int           `json:"skipped"`   // Skipped is the number of log line skipped by the method argument.
	Source    string        `json:"source"`    // Source is the source from which the log was read.
	Errors    []ErrorRecord `json:"errors"`    // Errors is the details of the unmatched record.
}

// Result represents serialized result data and metadata.
type Result struct {
	Data     []string `json:"data"`
	Metadata string   `json:"metadata"`
}

// controller is a controller for parsing logs. It parses the log line by line with the specified LineHandler and returns serialized data and aggregate metadata.
type controller func(input io.Reader, skipLines []int, patterns []*regexp.Regexp, handler LineHandler) ([]string, *Metadata, error)

// LineHandler is a type for functions that process matched lines.
type LineHandler func(matches []string, fields []string, index int) (string, error)

// MetadataHandler is a type for functions that process metadata.
type MetadataHandler func(metadata *Metadata) (string, error)

func parse(input io.Reader, skipLines []int, parser controller, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	data, metadata, err := parser(input, skipLines, patterns, lineHandler)
	if err != nil {
		return nil, err
	}
	return createResult(data, metadata, metadataHandler)
}

// ParseString parses the provided string input and returns a Result.
func parseString(input string, skipLines []int, parser controller, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	return parse(strings.NewReader(input), skipLines, parser, patterns, lineHandler, metadataHandler)
}

func parseFile(input string, skipLines []int, parser controller, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	data, metadata, err := parser(f, skipLines, patterns, lineHandler)
	if err != nil {
		return nil, err
	}
	metadata.Source = filepath.Base(input)
	return createResult(data, metadata, metadataHandler)
}

// ParseGzip parsers the content of a gzipped file and returns a Result.
func parseGzip(input string, skipLines []int, parser controller, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) (*Result, error) {
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
	data, metadata, err := parser(g, skipLines, patterns, lineHandler)
	if err != nil {
		return nil, err
	}
	metadata.Source = filepath.Base(input)
	return createResult(data, metadata, metadataHandler)
}

// ParseZipEntries parses the contents of entries in the zip archive that match the glob pattern and returns the result.
func parseZipEntries(input string, skipLines []int, globPattern string, parser controller, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) ([]*Result, error) {
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
		data, metadata, err := parser(e, skipLines, patterns, lineHandler)
		if err != nil {
			return nil, err
		}
		metadata.Source = f.Name
		out, err := createResult(data, metadata, metadataHandler)
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}

// parse reads from the input, processes each line and returns parsed data and metadata.
func parseRegex(input io.Reader, skipLines []int, patterns []*regexp.Regexp, handler LineHandler) ([]string, *Metadata, error) {
	if len(patterns) == 0 {
		return nil, nil, fmt.Errorf("cannot parse input: no patterns provided")
	}
	var data []string
	var errors []ErrorRecord
	i := 1
	matched := 0
	skipped := 0
	unmatched := 0
	m := make(map[int]bool, len(skipLines))
	for _, skipLine := range skipLines {
		m[skipLine] = true
	}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if _, ok := m[i]; ok {
			i++
			skipped++
			continue
		}
		r := scanner.Text()
		var matchedPattern *regexp.Regexp
		var matches []string
		for _, pattern := range patterns {
			matches = pattern.FindStringSubmatch(r)
			if matches != nil {
				matchedPattern = pattern
				break
			}
		}
		if matchedPattern == nil {
			errors = append(errors, ErrorRecord{Index: i, Record: r})
			unmatched++
			i++
			continue
		}
		names := matchedPattern.SubexpNames()
		labels := make([]string, 0, len(names)-1)
		values := make([]string, 0, len(names)-1)
		for j, name := range names[1:] {
			labels = append(labels, name)
			values = append(values, matches[j+1])
		}
		line, err := handler(values, labels, i)
		if err != nil {
			return nil, nil, err
		}
		data = append(data, line)
		matched++
		i++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("cannot read stream: %w", err)
	}
	metadata := &Metadata{
		Total:     i - 1,
		Matched:   matched,
		Skipped:   skipped,
		Unmatched: unmatched,
		Source:    "",
		Errors:    errors,
	}
	return data, metadata, nil
}

func parseLTSV(input io.Reader, skipLines []int, _ []*regexp.Regexp, handler LineHandler) ([]string, *Metadata, error) {
	var data []string
	var errors []ErrorRecord
	i := 1
	matched := 0
	skipped := 0
	unmatched := 0
	m := make(map[int]bool, len(skipLines))
	for _, skipLine := range skipLines {
		m[skipLine] = true
	}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if _, ok := m[i]; ok {
			i++
			skipped++
			continue
		}
		r := scanner.Text()
		fields := strings.Split(r, "\t")
		labels := make([]string, 0, len(fields))
		values := make([]string, 0, len(fields))
		for _, field := range fields {
			parts := strings.SplitN(field, ":", 2)
			if len(parts) != 2 {
				errors = append(errors, ErrorRecord{Index: i, Record: r})
				unmatched++
				i++
				continue
			}
			labels = append(labels, parts[0])
			values = append(values, parts[1])
		}
		line, err := handler(values, labels, i)
		if err != nil {
			return nil, nil, err
		}
		data = append(data, line)
		matched++
		i++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("cannot read stream: %w", err)
	}
	metadata := &Metadata{
		Total:     i - 1,
		Matched:   matched,
		Skipped:   skipped,
		Unmatched: unmatched,
		Source:    "",
		Errors:    errors,
	}
	return data, metadata, nil
}

// createResult is a helper function that constructs a Result from parsed data and metadata.
func createResult(data []string, metadata *Metadata, handler MetadataHandler) (*Result, error) {
	meta, err := handler(metadata)
	if err != nil {
		return nil, err
	}
	return &Result{
		Data:     data,
		Metadata: meta,
	}, nil
}
