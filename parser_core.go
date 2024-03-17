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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
)

type inputType int

// inputType defines the supported input formats for parsing.
const (
	inputTypeStream inputType = iota // indicates parsing from a stream (e.g., stdin).
	inputTypeString                  // indicates parsing directly from a string.
	inputTypeFile                    // indicates parsing from a file on disk.
	inputTypeGzip                    // indicates parsing from a gzip-compressed file.
	inputTypeZip                     // indicates parsing from a file within a zip archive.
)

// common error messages
const (
	parseError        = "cannot parse input"
	resultError       = "invalid parsing result"
	globPatternError  = "invalid glob pattern"
	regexPatternError = "invalid regex pattern"
	emptyPathError    = "empty path detected"
	openFileError     = "cannot open file"
	filterError       = "cannot evaluate filter expressions"
	operatorError     = "unknown operator"
)

// Parser interface defines methods for parsing log data from various sources.
// Basically used internally to implement RegexParser and LTSVParser.
type Parser interface {
	Parse(reader io.Reader) (*Result, error)
	ParseString(s string) (*Result, error)
	ParseFile(filePath string) (*Result, error)
	ParseGzip(gzipPath string) (*Result, error)
	ParseZipEntries(zipPath, globPattern string) (*Result, error)
}

// Option defines the parser settings.
// Each field is used to customize the output.
type Option struct {
	Labels       []string    // specify fields to output by label name
	Filters      []string    // conditional expression for output log lines
	SkipLines    []int       // line numbers to exclude from output (not index)
	Prefix       bool        // whether to prefix the output lines or not
	UnmatchLines bool        // whether to output unmatched lines as raw logs or not
	LineNumber   bool        // whether to add line numbers or not
	LineHandler  LineHandler // handler function to convert log lines
}

// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line number, and returns processed string data.
type lineDecoder func(line string, patterns []*regexp.Regexp) ([]string, []string, error)

// lineFilter is a function type that provides a filter function applied to log lines.
type lineFilter func(v string) (bool, error)

// LineHandler is a function type that processes each matched line.
// It takes the matches, their corresponding fields, and the line number, and returns processed string data.
type LineHandler func(labels []string, values []string, isFirst bool) (string, error)

// parse orchestrates the parsing process, applying keyword filters and regular expression patterns to log data from an io.Reader.
// It supports dynamic handling of line processing, error collection, and pattern matching for efficient log analysis.
// This function is used as an internal process of the Parse method.
func parse(ctx context.Context, input io.Reader, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	r, err := parser(ctx, input, output, patterns, decoder, opt)
	if err != nil {
		return nil, err
	}
	r.inputType = inputTypeStream
	return r, nil
}

// parseString is a convenience method for parsing log data directly from a string.
// It applies the same processing as parse, tailored for situations where log data is available as a string.
// This function is used as an internal process of the ParseString method.
func parseString(ctx context.Context, s string, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	r, err := parser(ctx, strings.NewReader(s), output, patterns, decoder, opt)
	if err != nil {
		return nil, err
	}
	r.inputType = inputTypeString
	return r, nil
}

// parseFile opens and processes log data from a file, applying the specified patterns and handlers.
// It handles file opening/closing and applies the same log processing logic as parse.
// This function is used as an internal process of the ParseFile method.
func parseFile(ctx context.Context, filePath string, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	f, cleanup, err := handleFile(filePath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	r, err := parser(ctx, f, output, patterns, decoder, opt)
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
func parseGzip(ctx context.Context, gzipPath string, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	g, cleanup, err := handleGzip(gzipPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	r, err := parser(ctx, g, output, patterns, decoder, opt)
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
func parseZipEntries(ctx context.Context, zipPath, globPattern string, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	result := Result{Errors: make([]Errors, 0)}
	err := handleZipEntries(zipPath, globPattern, func(f *zip.File) error {
		e, err := f.Open()
		if err != nil {
			return fmt.Errorf("%s: %w", openFileError, err)
		}
		defer e.Close()
		r, err := parser(ctx, e, output, patterns, decoder, opt)
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

// parser is the core logic of this module. It processes an input stream line by line against a set of regular expression patterns,
// filters, and additional processing options. It applies specified filters, handles matched lines using a custom line handler, and
// writes results to an output stream.
func parser(ctx context.Context, input io.Reader, output io.Writer, patterns []*regexp.Regexp, decoder lineDecoder, opt Option) (*Result, error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	start := time.Now()
	r := &Result{Errors: make([]Errors, 0)}
	i := 0
	m := applySkipLines(opt.SkipLines)
	isFirst := true
	mpref := "[ PROCESSED ] "
	upref := "[ UNMATCHED ] "
	if isatty.IsTerminal(os.Stdout.Fd()) {
		mpref = "\033[1;32m" + mpref + "\033[0m"
		upref = "\033[1;31m" + upref + "\033[0m"
	}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return r, ctx.Err()
		default:
			i++
			if _, ok := m[i]; ok {
				r.Skipped++
				continue
			}
			raw := scanner.Text()
			praw := raw
			if opt.Prefix {
				praw = upref + raw
			}
			ls, vs, err := decoder(raw, patterns)
			if err != nil {
				if strings.Contains(err.Error(), "no pattern provided") {
					return nil, err
				}
				if opt.UnmatchLines {
					if _, err := fmt.Fprintln(output, praw); err != nil {
						return nil, err
					}
				}
				r.Errors = append(r.Errors, Errors{LineNumber: i, Line: raw})
				r.Unmatched++
				continue
			}
			f, err := applyFilter(ls, vs, opt.Filters)
			if err != nil {
				return nil, err
			}
			if !f {
				r.Excluded++
				continue
			}
			if len(opt.Labels) > 0 {
				ls, vs = selectLabels(opt.Labels, ls, vs)
			}
			if opt.LineNumber {
				ls, vs = addLineNumber(ls, vs, i)
			}
			line, err := opt.LineHandler(ls, vs, isFirst)
			if err != nil {
				return nil, err
			}
			if opt.Prefix {
				line = applyPrefix(line, mpref)
			}
			if _, err := fmt.Fprintln(output, line); err != nil {
				return nil, err
			}
			r.Matched++
			isFirst = false
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
		return nil, nil, fmt.Errorf("%s: no pattern provided", parseError)
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(line)
		if matches != nil {
			return pattern.SubexpNames()[1:], matches[1:], nil
		}
	}
	return nil, nil, fmt.Errorf("%s: no matching pattern for line: \"%s\"", parseError, line)
}

// ltsvDecoder parses a string formatted in Labeled Tab-separated Values (LTSV)
// format. It splits the string into fields based on tabs and then further
// splits each field into labels and values. Returns an error for invalid fields.
func ltsvLineDecoder(line string, _ []*regexp.Regexp) ([]string, []string, error) {
	fields := strings.Split(line, "\t")
	ls := make([]string, 0, len(fields))
	vs := make([]string, 0, len(fields))
	for _, field := range fields {
		token := strings.SplitN(field, ":", 2)
		if len(token) != 2 {
			return nil, nil, fmt.Errorf("%s: invalid field: \"%s\"", parseError, field)
		}
		ls = append(ls, token[0])
		vs = append(vs, token[1])
	}
	return ls, vs, nil
}

// selectLabels filters the given labels and values based on a list of target labels.
func selectLabels(targets, labels, values []string) ([]string, []string) {
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

// addLineNumber prepends the line number to labels and values.
func addLineNumber(labels []string, values []string, lineNumber int) ([]string, []string) {
	return append([]string{"no"}, labels...), append([]string{strconv.Itoa(lineNumber)}, values...)
}

// applySkipLines generates a map indicating which line numbers should be skipped during parsing.
// It takes a slice of line numbers to skip and returns a map with these line numbers as keys.
func applySkipLines(skipLines []int) map[int]struct{} {
	m := make(map[int]struct{}, len(skipLines))
	for _, skipLine := range skipLines {
		m[skipLine] = struct{}{}
	}
	return m
}

// applyPrefix sets the play fix for log lines.
func applyPrefix(line, prefix string) string {
	b := &strings.Builder{}
	lines := strings.Split(line, "\n")
	for i, l := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(prefix)
		b.WriteString(l)
	}
	return b.String()
}

// applyFilter evaluates a filter expression passed as a string and controls
// whether or not log lines are output according to the result.
func applyFilter(labels, values, filters []string) (bool, error) {
	m, err := getFilter(labels, filters)
	if err != nil {
		return false, err
	}
	for i, label := range labels {
		if filter, ok := m[label]; ok {
			f, err := filter(values[i])
			if err != nil {
				return false, err
			}
			if !f {
				return false, nil
			}
		}
	}
	return true, nil
}

// getFilter creates a map of lineFilter functions based on the provided filters and labels.
// Each filter is parsed into a label, operator, and value, and an appropriate lineFilter
// function is created to match lines accordingly. This function validates the syntax of
// filter expressions and ensures that each label used in the filters is included in the
// provided labels list.
func getFilter(labels, filters []string) (map[string]lineFilter, error) {
	m := map[string]lineFilter{}
	for _, filter := range filters {
		token := strings.SplitN(filter, " ", 3)
		if len(token) < 3 {
			return nil, fmt.Errorf("%s: \"%s\": invalid syntax", filterError, filter)
		}
		label, operator, value := token[0], token[1], token[2]
		if !slices.Contains(labels, label) {
			return nil, fmt.Errorf("%s: \"%s\": invalid field name", filterError, label)
		}
		switch operator {
		case "==", "!=", "==*", "!=*":
			f, err := getStringFilter(operator, value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", filterError, err)
			}
			m[label] = f
		case "=~", "!~", "=~*", "!~*":
			f, err := getRegexFilter(operator, value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", filterError, err)
			}
			m[label] = f
		case ">", ">=", "<", "<=":
			f, err := getNumericFilter(operator, value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", filterError, err)
			}
			m[label] = f
		default:
			return nil, fmt.Errorf("%s: \"%s\"", operatorError, operator)
		}
	}
	return m, nil
}

// getStringFilter returns a lineFilter function for string comparison based on the specified
// operator and value. Supported operators are "==", "!=", "==*" (case-insensitive equal), and
// "!=*" (case-insensitive not equal).
func getStringFilter(operator, value string) (lineFilter, error) {
	switch operator {
	case "==":
		return func(v string) (bool, error) { return v == value, nil }, nil
	case "!=":
		return func(v string) (bool, error) { return v != value, nil }, nil
	case "==*":
		return func(v string) (bool, error) { return strings.EqualFold(v, value), nil }, nil
	case "!=*":
		return func(v string) (bool, error) { return !strings.EqualFold(v, value), nil }, nil
	default:
		return nil, fmt.Errorf("%s: \"%s\"", operatorError, operator)
	}
}

// getNumericFilter returns a lineFilter function for numeric comparison based on the specified
// operator and value. The function compares the numeric value of a string against the provided
// value using the specified operator. Supported operators are ">", ">=", "<", and "<=".
func getNumericFilter(operator, value string) (lineFilter, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	switch operator {
	case ">":
		return func(v string) (bool, error) {
			val, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false, err
			}
			return val > f, nil
		}, nil
	case ">=":
		return func(v string) (bool, error) {
			val, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false, err
			}
			return val >= f, nil
		}, nil
	case "<":
		return func(v string) (bool, error) {
			val, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false, err
			}
			return val < f, nil
		}, nil
	case "<=":
		return func(v string) (bool, error) {
			val, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return false, err
			}
			return val <= f, nil
		}, nil
	default:
		return nil, fmt.Errorf("%s: \"%s\"", operatorError, operator)
	}
}

// getRegexFilter returns a lineFilter function that matches a string against a regular expression
// pattern. The function supports both standard and case-insensitive matches, specified by the
// "=~" and "=~*" (or "!~" and "!~*" for negation) operators, respectively.
func getRegexFilter(operator, value string) (lineFilter, error) {
	switch operator {
	case "=~*", "!~*":
		value = "(?i)" + value
	}
	re, err := regexp.Compile(value)
	if err != nil {
		return nil, err
	}
	switch operator {
	case "=~", "=~*":
		return func(v string) (bool, error) { return re.MatchString(v), nil }, nil
	case "!~", "!~*":
		return func(v string) (bool, error) { return !re.MatchString(v), nil }, nil
	default:
		return nil, fmt.Errorf("%s: \"%s\"", operatorError, operator)
	}
}

// handleFile opens a file for reading, ensuring it is properly closed after processing.
// It abstracts file handling, providing a clean and reusable way to work with file resources.
func handleFile(filePath string) (*os.File, func(), error) {
	if filePath == "" {
		return nil, nil, fmt.Errorf(emptyPathError)
	}
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", openFileError, err)
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
		return nil, nil, fmt.Errorf(emptyPathError)
	}
	f, err := os.Open(filepath.Clean(gzipPath))
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", openFileError, err)
	}
	g, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, err
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
		return fmt.Errorf(emptyPathError)
	}
	z, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("%s: %w", openFileError, err)
	}
	defer z.Close()
	for _, f := range z.File {
		matched, err := filepath.Match(globPattern, f.Name)
		if err != nil {
			return fmt.Errorf("%s: %w", globPatternError, err)
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
