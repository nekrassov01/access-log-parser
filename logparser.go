// Package logparser provides utilities for parsing access logs and converting them to structured formats.
package logparser

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
const Version = "0.0.6"

// Parser is a struct that defines how to parse the access log.
// Patterns represents a list of regular expression patterns used for matching log lines,
// which are matched in order. Each field must have a named capture group.
// LineHandler is a custom function that processes the matched log lines.
// If not provided at instantiation, DefaultLineHandler will be used.
// MetadataHandler is a custom function that processes and formats metadata.
// If not provided at instantiation, DefaultMetadataHandler will be used.
type Parser struct {
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

// Result represents serialized result data and metadata.
type Result struct {
	Data     []string `json:"data"`
	Metadata string   `json:"metadata"`
}

// LineHandler is a type for functions that process matched lines.
type LineHandler func(matches []string, fields []string, index int) (string, error)

// MetadataHandler is a type for functions that process metadata.
type MetadataHandler func(metadata *Metadata) (string, error)

// Option provides an option to override the handler behavior of Parser.
type Option func(*Parser)

// WithLineHandler overrides the behavior of DefaultLineHandler.
func WithLineHandler(handler LineHandler) Option {
	return func(p *Parser) {
		p.LineHandler = handler
	}
}

// WithMetadataHandler overrides the behavior of DefaultMetadataHandler.
func WithMetadataHandler(handler MetadataHandler) Option {
	return func(p *Parser) {
		p.MetadataHandler = handler
	}
}

// NewParser creates a new parser with the specified options.
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		LineHandler:     DefaultLineHandler,
		MetadataHandler: DefaultMetadataHandler,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// AddPattern sets the regex pattern for Parser. Parser holds slice of patterns,
// so use AddPatterns if you want to set multiple patterns at once.
func (p *Parser) AddPattern(pattern *regexp.Regexp) error {
	if len(pattern.SubexpNames()) <= 1 {
		return fmt.Errorf("invalid pattern detected: capture group not found")
	}
	for j, name := range pattern.SubexpNames() {
		if j != 0 && name == "" {
			return fmt.Errorf("invalid pattern detected: non-named capture group detected")
		}
	}
	p.Patterns = append(p.Patterns, pattern)
	return nil
}

// AddPatterns sets multiple regex patterns at once.
// If an error is detected, all patterns are cleared.
func (p *Parser) AddPatterns(patterns []*regexp.Regexp) error {
	for _, pattern := range patterns {
		if err := p.AddPattern(pattern); err != nil {
			p.Patterns = nil
			return err
		}
	}
	return nil
}

// parse reads from the input, processes each line and returns parsed data and metadata.
func (p *Parser) parse(input io.Reader, skipLines []int) ([]string, *Metadata, error) {
	if len(p.Patterns) == 0 {
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
		for _, pattern := range p.Patterns {
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
		fields := make([]string, 0, len(names)-1)
		values := make([]string, 0, len(names)-1)
		for j, name := range names[1:] {
			fields = append(fields, name)
			values = append(values, matches[j+1])
		}
		line, err := p.LineHandler(values, fields, i)
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

// Parse parses the provided io.Reader and returns a Result.
// Use this if you want to use an abstracted Reader.
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
	for i, match := range matches {
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
