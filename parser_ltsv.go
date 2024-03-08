package parser

import (
	"context"
	"io"
)

var _ Parser = (*LTSVParser)(nil)

// LTSVParser implements the Parser interface for parsing logs in LTSV (Labeled Tab-separated Values) format.
// It allows customization of line handling for LTSV formatted data.
type LTSVParser struct {
	ctx             context.Context
	writer          io.Writer
	labels          []string
	filters         []string
	skipLines       []int
	hasPrefix       bool
	hasUnmatchLines bool
	hasLineNumber   bool
	lineDecoder     lineDecoder
	lineHandler     LineHandler
}

// NewLTSVParser initializes a new LTSVParser with default handlers for line decoding, line handling.
// This parser is specifically tailored for LTSV formatted log data.
func NewLTSVParser(ctx context.Context, writer io.Writer) *LTSVParser {
	return &LTSVParser{
		ctx:         ctx,
		writer:      writer,
		lineDecoder: ltsvLineDecoder,
		lineHandler: JSONLineHandler,
	}
}

// SelectLabels sets the label names to be output.
// Only selected fields can be returned.
func (p *LTSVParser) SelectLabels(labels []string) *LTSVParser {
	p.labels = labels
	return p
}

// SetSkipLines specifies the line numbers to skip.
// This is useful if you want to exclude headers and footers.
func (p *LTSVParser) SetSkipLines(skipLines []int) *LTSVParser {
	p.skipLines = skipLines
	return p
}

// EnablePrefix determines whether log lines are prefixed with [ PROCESSED | UNMATCHED ] etc.
// It is intended to be used with EnableUnmatchLines. Default is false.
func (p *LTSVParser) EnablePrefix(has bool) *LTSVParser {
	p.hasPrefix = has
	return p
}

// EnableUnmatchLines determines whether lines that did not match should also be output as raw logs.
func (p *LTSVParser) EnableUnmatchLines(has bool) *LTSVParser {
	p.hasUnmatchLines = has
	return p
}

// EnableLineNumber determines whether line numbers are set.
func (p *LTSVParser) EnableLineNumber(has bool) *LTSVParser {
	p.hasLineNumber = has
	return p
}

// SetFilters sets the filter expression for log lines as a string.
// "*" denotes case-insensitive.
//
//	> >= == <= < (arithmetic (float64))
//	== ==* != !=* (string comparison (string))
//	=~ !~ =~* !~* (regular expression (string))
func (p *LTSVParser) SetFilters(filters []string) *LTSVParser {
	p.filters = filters
	return p
}

// SetLineHandler sets the function responsible for processing each line of log data.
func (p *LTSVParser) SetLineHandler(handler LineHandler) *LTSVParser {
	if handler == nil {
		return p
	}
	p.lineHandler = handler
	return p
}

// Parse processes log data from an io.Reader, applying the configured line handlers.
// This method supports context cancellation, prefixing of lines, and exclusion of specific lines.
func (p *LTSVParser) Parse(reader io.Reader) (*Result, error) {
	return parse(p.ctx, reader, p.writer, nil, p.labels, p.filters, p.skipLines, p.hasPrefix, p.hasUnmatchLines, p.hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseString processes a log string directly, applying configured skip lines and line number handling.
// It's designed for quick parsing of a single LTSV formatted log string.
func (p *LTSVParser) ParseString(s string) (*Result, error) {
	return parseString(p.ctx, s, p.writer, nil, p.labels, p.filters, p.skipLines, p.hasPrefix, p.hasUnmatchLines, p.hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseFile reads and parses log data from a file, leveraging the configured patterns and handlers.
// This method simplifies file-based LTSV log parsing with automatic line processing.
func (p *LTSVParser) ParseFile(filePath string) (*Result, error) {
	return parseFile(p.ctx, filePath, p.writer, nil, p.labels, p.filters, p.skipLines, p.hasPrefix, p.hasUnmatchLines, p.hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseGzip processes gzip-compressed log data, extending the parser's capabilities to compressed LTSV logs.
// It applies skip lines and line number handling as configured for gzip-compressed files.
func (p *LTSVParser) ParseGzip(gzipPath string) (*Result, error) {
	return parseGzip(p.ctx, gzipPath, p.writer, nil, p.labels, p.filters, p.skipLines, p.hasPrefix, p.hasUnmatchLines, p.hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseZipEntries processes log data within zip archive entries, applying skip lines, line number handling,
// and optional glob pattern matching. This method is ideal for batch processing of LTSV logs in zip files.
func (p *LTSVParser) ParseZipEntries(zipPath, globPattern string) (*Result, error) {
	return parseZipEntries(p.ctx, zipPath, globPattern, p.writer, nil, p.labels, p.filters, p.skipLines, p.hasPrefix, p.hasUnmatchLines, p.hasLineNumber, p.lineDecoder, p.lineHandler)
}
