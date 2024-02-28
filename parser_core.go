// Package parser provides utilities for parsing various types of logs (plain text, gzip, zip)
// and converting them into structured formats such as JSON or LTSV. It supports pattern matching,
// result extraction, and error handling to facilitate log analysis and data extraction.
package parser

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type inputType int

// inputType defines the supported input formats for parsing.
const (
	inputTypeStream inputType = iota // Indicates parsing from a stream (e.g., stdin).
	inputTypeString                  // Indicates parsing directly from a string.
	inputTypeFile                    // Indicates parsing from a file on disk.
	inputTypeGzip                    // Indicates parsing from a gzip-compressed file.
	inputTypeZip                     // Indicates parsing from a file within a zip archive.
)

// Parser interface defines methods for parsing log data from various sources.
// Implementations must handle keyword filtering, label extraction, and processing of lines based
// on provided patterns.
//   - SetLineHandler sets the function to handle each parsed line.
//   - Parse processes log data from io.Reader and applies keyword filtering and pattern matching.
//     Use this method if you want to process real-time streaming from stdin or other sources;
//     upon interruption by an Interruput signal, it exits gracefully, retaining aggregate results.
//   - ParseString processes log data from a string.
//   - ParseFile processes log data from a file on disk.
//   - ParseGzip processes log data from a gzip-compressed file.
//   - ParseZipEntries processes log data matching the glob pattern from files in a zip archive.
//     Aggregate results are merged, and options such as keyword filtering and exclusion
//     by line number are applied in the same way to the target files.
type Parser interface {
	SetLineHandler(LineHandler)
	Parse(ctx context.Context, reader io.Reader, keywords, labels []string, hasPrefix, disableUnmatch bool) (*Result, error)
	ParseString(s string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error)
	ParseFile(filePath string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error)
	ParseGzip(gzipPath string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error)
	ParseZipEntries(zipPath, globPattern string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error)
}

// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line index, and returns processed string data.
type lineDecoder func(line string, patterns []*regexp.Regexp) ([]string, []string, error)

// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line index, and returns processed string data.
type LineHandler func(labels []string, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error)

// parse orchestrates the parsing process, applying keyword filters and regular expression patterns to log data from an io.Reader.
// It supports dynamic handling of line processing, error collection, and pattern matching for efficient log analysis.
// This function is used as an internal process of the Parse method.
func parse(ctx context.Context, input io.Reader, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, hasPrefix, disableUnmatch bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	r, err := streamer(ctx, input, output, patterns, keywords, labels, hasPrefix, disableUnmatch, decoder, handler)
	if err != nil {
		return nil, err
	}
	r.inputType = inputTypeStream
	return r, nil
}

// parseString is a convenience method for parsing log data directly from a string.
// It applies the same processing as parse, tailored for situations where log data is available as a string.
// This function is used as an internal process of the ParseString method.
func parseString(s string, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, skipLines []int, hasLineNumber bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	r, err := parser(strings.NewReader(s), output, patterns, keywords, labels, skipLines, hasLineNumber, decoder, handler)
	if err != nil {
		return nil, err
	}
	r.inputType = inputTypeString
	return r, nil
}

// parseFile opens and processes log data from a file, applying the specified patterns and handlers.
// It handles file opening/closing and applies the same log processing logic as parse.
// This function is used as an internal process of the ParseFile method.
func parseFile(filePath string, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, skipLines []int, hasLineNumber bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	f, cleanup, err := handleFile(filePath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	r, err := parser(f, output, patterns, keywords, labels, skipLines, hasLineNumber, decoder, handler)
	if err != nil {
		return nil, err
	}
	r.Source = filepath.Base(filePath)
	r.inputType = inputTypeFile
	return r, nil
}

// parseGzip opens a gzip-compressed log file and processes its contents.
// It allows for the direct parsing of compressed logs, applying the specified patterns and handlers.
// This function is used as an internal process of the ParseGzip method.
func parseGzip(gzipPath string, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, skipLines []int, hasLineNumber bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	g, cleanup, err := handleGzip(gzipPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	r, err := parser(g, output, patterns, keywords, labels, skipLines, hasLineNumber, decoder, handler)
	if err != nil {
		return nil, err
	}
	r.Source = filepath.Base(gzipPath)
	r.inputType = inputTypeGzip
	return r, nil
}

// parseZipEntries processes log entries within a zip archive, filtering files based on a glob pattern.
// It enables the parsing of multiple log files contained within a single archive.
// This function is used as an internal process of the ParseZipEntries method.
func parseZipEntries(zipPath, globPattern string, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, skipLines []int, hasLineNumber bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	result := Result{Errors: make([]Errors, 0)}
	err := handleZipEntries(zipPath, globPattern, func(f *zip.File) error {
		e, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open zip file entry: %w", err)
		}
		defer e.Close()
		r, err := parser(e, output, patterns, keywords, labels, skipLines, hasLineNumber, decoder, handler)
		if err != nil {
			return err
		}
		for i := range r.Errors {
			r.Errors[i].Entry = f.Name
		}
		result.Total += r.Total
		result.Matched += r.Matched
		result.Unmatched += r.Unmatched
		result.Excluded += r.Excluded
		result.Skipped += r.Skipped
		result.ElapsedTime += r.ElapsedTime
		result.Source = filepath.Base(zipPath)
		result.ZipEntries = append(result.ZipEntries, f.Name)
		result.Errors = append(result.Errors, r.Errors...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	result.inputType = inputTypeZip
	return &result, nil
}

// The parser processes log data from io.Reader and provides the core logic for parsing log lines by the decoder
// and applying transformations by the handler. The processed log lines are sequentially output to io.Writer,
// which returns an aggregate result after all lines have been processed.
func parser(input io.Reader, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, skipLines []int, hasLineNumber bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	start := time.Now()
	r := &Result{Errors: make([]Errors, 0)}
	i := 0
	m := skipLineMap(skipLines)
	isFirst := true
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		i++
		if _, ok := m[i]; ok {
			r.Skipped++
			continue
		}
		raw := scanner.Text()
		if !hasKeyword(raw, keywords) {
			r.Excluded++
			continue
		}
		ls, vs, err := decoder(raw, patterns)
		if err != nil {
			if strings.Contains(err.Error(), "no pattern provided") {
				return nil, err
			}
			r.Errors = append(r.Errors, Errors{LineNumber: i, Line: raw})
			r.Unmatched++
			continue
		}
		ls, vs = selectColumns(labels, ls, vs)
		line, err := handler(ls, vs, i, hasLineNumber, isFirst)
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return nil, err
		}
		r.Matched++
		isFirst = false
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	r.Total = i
	r.ElapsedTime = time.Since(start)
	return r, nil
}

// streamer is a streaming-specific function of parser. It reads log data in a streaming fashion
// and applies pattern matching and row processing in real time.
// Supports context-based cancellation to allow graceful termination of log processing.
func streamer(ctx context.Context, input io.Reader, output io.Writer, patterns []*regexp.Regexp, keywords, labels []string, hasPrefix, disableUnmatch bool, decoder lineDecoder, handler LineHandler) (*Result, error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	start := time.Now()
	r := &Result{Errors: make([]Errors, 0)}
	i := 0
	mpref := "\033[1;32m[ PROCESSED ] \033[0m"
	upref := "\033[1;31m[ UNMATCHED ] \033[0m"
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return r, ctx.Err()
		default:
			i++
			raw := scanner.Text()
			if !hasKeyword(raw, keywords) {
				r.Excluded++
				continue
			}
			praw := raw
			if hasPrefix {
				praw = upref + raw
			}
			ls, vs, err := decoder(raw, patterns)
			if err != nil {
				if strings.Contains(err.Error(), "no pattern provided") {
					return nil, err
				}
				if !disableUnmatch {
					if _, err := fmt.Fprintln(output, praw); err != nil {
						return nil, err
					}
				}
				r.Errors = append(r.Errors, Errors{LineNumber: i, Line: raw})
				r.Unmatched++
				continue
			}
			ls, vs = selectColumns(labels, ls, vs)
			line, err := handler(ls, vs, i, false, false)
			if err != nil {
				return nil, err
			}
			if hasPrefix {
				line = mpref + line
			}
			if _, err := fmt.Fprintln(output, line); err != nil {
				return nil, err
			}
			r.Matched++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	r.Total = i
	r.ElapsedTime = time.Since(start)
	return r, nil
}

// regexLineDecoder applies regular expression patterns to a given string and
// extracts matching groups. It returns slices of labels and values extracted
// from the string. If no pattern matches, it returns an error.
func regexLineDecoder(line string, patterns []*regexp.Regexp) ([]string, []string, error) {
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
func ltsvLineDecoder(line string, _ []*regexp.Regexp) ([]string, []string, error) {
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

// handleFile opens a file for reading, ensuring it is properly closed after processing.
// It abstracts file handling, providing a clean and reusable way to work with file resources.
func handleFile(filePath string) (*os.File, func(), error) {
	if filePath == "" {
		return nil, nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open file: %w", err)
	}
	cleanup := func() {
		f.Close()
	}
	return f, cleanup, nil
}

// handleGzip opens a gzip-compressed file and prepares it for reading, handling decompression transparently.
// It simplifies working with gzip files, abstracting away the details of decompression.
func handleGzip(gzipPath string) (*gzip.Reader, func(), error) {
	if gzipPath == "" {
		return nil, nil, fmt.Errorf("empty path detected")
	}
	f, err := os.Open(filepath.Clean(gzipPath))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open file: %w", err)
	}
	g, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("cannot create gzip reader for %s: %w", gzipPath, err)
	}
	cleanup := func() {
		g.Close()
		f.Close()
	}
	return g, cleanup, nil
}

// handleZipEntries iterates over entries in a zip file, applying a provided function to each matching entry.
// It supports glob pattern matching for entry names, enabling selective processing of zip contents.
func handleZipEntries(zipPath string, globPattern string, fn func(f *zip.File) error) error {
	if zipPath == "" {
		return fmt.Errorf("empty path detected")
	}
	z, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("cannot open zip file: %w", err)
	}
	defer z.Close()
	for _, f := range z.File {
		matched, err := filepath.Match(globPattern, f.Name)
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}
		if !matched {
			continue
		}
		if err := fn(f); err != nil {
			return err
		}
	}
	return nil
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

// hasKeyword checks if the given s contains any of the specified keywords.
func hasKeyword(s string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

// selectColumns filters the given labels and values based on a list of target labels.
func selectColumns(targets, labels, values []string) ([]string, []string) {
	if len(targets) == 0 {
		return labels, values
	}
	m := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		m[target] = struct{}{}
	}
	ls := make([]string, 0, len(targets))
	vs := make([]string, 0, len(targets))
	for j, label := range labels {
		if _, ok := m[label]; ok {
			ls = append(ls, label)
			vs = append(vs, values[j])
		}
	}
	return ls, vs
}
