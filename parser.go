// Package parser provides utilities for parsing access logs and converting them to structured formats.
package parser

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Version of access-log-parser.
const Version = "0.0.4"

// Parser is a structure that defines how to parse the access log.
// Fields represents the field names of the log entries.
// Patterns represents a list of regex patterns used for matching log lines, which are matched in order.
// LineHandler is a custom function that processes the matched log lines.
// If not provided at instantiation, DefaultLineHandler is used.
// MetadataHandler is a custom function that processes and formats metadata.
// If not provided at instantiation, DefaultMetadataHandler is used.
type Parser struct {
	Fields          []string         `json:"fields"`
	Patterns        []*regexp.Regexp `json:"patterns"`
	LineHandler     LineHandler      `json:"-"`
	MetadataHandler MetadataHandler  `json:"-"`
}

// ErrorRecord represents a record that didn't match any of the provided patterns.
// Such records are stored in this structure to prevent data loss and for further inspection.
// Index is the index of the input line that didn't match.
// Record is the actual input line that was unmatched.
type ErrorRecord struct {
	Index  int    `json:"index"`
	Record string `json:"record"`
}

// Metadata represents metadata about the parsed records.
// Total is the total number of log entries processed.
// Matched is the number of log entries that successfully matched a pattern.
// Unmatched is the number of log entries that did not match any pattern.
// Skipped is the number of log entries skipped by the method argument.
// Source is the source from which the log was read.
// Errors is the details of the unmatched record.
type Metadata struct {
	Total     int           `json:"total"`
	Matched   int           `json:"matched"`
	Unmatched int           `json:"unmatched"`
	Skipped   int           `json:"skipped"`
	Source    string        `json:"source"`
	Errors    []ErrorRecord `json:"errors"`
}

// Result represents the parsed log data and metadata.
type Result struct {
	Data     []string `json:"data"`
	Metadata string   `json:"metadata"`
}

// LineHandler is a type for functions that process matched lines.
type LineHandler func(matches []string, fields []string, index int) (string, error)

// MetadataHandler is a type for functions that process metadata.
type MetadataHandler func(metadata *Metadata) (string, error)

// New creates a new parser with the specified configurations.
func New(fields []string, patterns []*regexp.Regexp, lineHandler LineHandler, metadataHandler MetadataHandler) *Parser {
	if lineHandler == nil {
		lineHandler = DefaultLineHandler
	}
	if metadataHandler == nil {
		metadataHandler = DefaultMetadataHandler
	}
	return &Parser{
		Fields:          fields,
		Patterns:        patterns,
		LineHandler:     lineHandler,
		MetadataHandler: metadataHandler,
	}
}

// parse reads from the input, processes each line and returns parsed data and metadata.
func (p *Parser) parse(input io.Reader, skipLines []int) ([]string, *Metadata, error) {
	var data []string
	var errors []ErrorRecord
	i := 1
	matched := 0
	skipped := 0
	unmatched := 0
	m := make(map[int]bool)
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
		var matches []string
		r := scanner.Text()
		for _, pattern := range p.Patterns {
			if matches = pattern.FindStringSubmatch(r); matches != nil {
				break
			}
		}
		if matches == nil {
			errors = append(errors, ErrorRecord{Index: i, Record: r})
			unmatched++
		} else {
			line, err := p.LineHandler(matches, p.Fields, i)
			if err != nil {
				return nil, nil, err
			}
			data = append(data, line)
			matched++
		}
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

// Parse parses the provided io.Reader input and returns a Result.
func (p *Parser) Parse(input io.Reader, skipLines []int) (*Result, error) {
	data, meta, err := p.parse(input, skipLines)
	if err != nil {
		return nil, err
	}
	return p.createResult(data, meta)
}

// ParseString parses the provided string input and returns a Result.
func (p *Parser) ParseString(input string, skipLines []int) (*Result, error) {
	return p.Parse(strings.NewReader(input), skipLines)
}

// ParseFile parsers the content of a file and returns a Result.
func (p *Parser) ParseFile(input string, skipLines []int) (*Result, error) {
	if input == "" {
		return nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(input))
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	data, meta, err := p.parse(f, skipLines)
	if err != nil {
		return nil, err
	}
	meta.Source = filepath.Base(input)
	return p.createResult(data, meta)
}

// ParseGzip parsers the content of a gzipped file and returns a Result.
func (p *Parser) ParseGzip(input string, skipLines []int) (*Result, error) {
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
	data, meta, err := p.parse(g, skipLines)
	if err != nil {
		return nil, err
	}
	meta.Source = filepath.Base(input)
	return p.createResult(data, meta)
}

// ParseZipEntries parses the contents of entries in the zip archive that match the glob pattern and returns the result.
func (p *Parser) ParseZipEntries(input string, skipLines []int, globPattern string) ([]*Result, error) {
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
		data, meta, err := p.parse(e, skipLines)
		if err != nil {
			return nil, err
		}
		meta.Source = f.Name
		out, err := p.createResult(data, meta)
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}

// createResult is a helper function that constructs a Result from parsed data and metadata.
func (p *Parser) createResult(data []string, metadata *Metadata) (*Result, error) {
	meta, err := p.MetadataHandler(metadata)
	if err != nil {
		return nil, err
	}
	return &Result{
		Data:     data,
		Metadata: meta,
	}, nil
}

// DefaultLineHandler is the default handler that converts matched log entries to NDJSON format.
// It is used if no specific LineHandler is provided when the Parser is instantiated.
func DefaultLineHandler(matches []string, fields []string, index int) (string, error) {
	var builder strings.Builder
	_, err := builder.WriteString(fmt.Sprintf("{\"index\":%d", index))
	if err != nil {
		return "", fmt.Errorf("cannot use builder to write strings: %w", err)
	}
	for i, match := range matches[1:] {
		if i < len(fields) {
			b, err := json.Marshal(match)
			if err != nil {
				return "", fmt.Errorf("cannot marshal matched string \"%s\" as json: %w", match, err)
			}
			_, err = builder.WriteString(fmt.Sprintf(",\"%s\":%s", fields[i], b))
			if err != nil {
				return "", fmt.Errorf("cannot use builder to write strings: %w", err)
			}
		}
	}
	builder.WriteString("}")
	return builder.String(), nil
}

// DefaultMetadataHandler is the default handler that converts parsed log metadata to NDJSON format.
// It is used if no specific MetadataHandler is provided when the Parser is instantiated.
func DefaultMetadataHandler(m *Metadata) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("cannot marshal result as json: %w", err)
	}
	return string(b), nil
}
