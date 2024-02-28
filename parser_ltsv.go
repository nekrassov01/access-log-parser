package parser

import (
	"context"
	"io"
)

var _ Parser = (*LTSVParser)(nil)

// LTSVParser implements the Parser interface for parsing logs in LTSV (Labeled Tab-separated Values) format.
// It allows customization of line and metadata handling for LTSV formatted data.
type LTSVParser struct {
	writer      io.Writer
	lineDecoder lineDecoder
	lineHandler LineHandler
}

// NewLTSVParser initializes a new LTSVParser with default handlers for line decoding, line handling,
// and metadata handling. This parser is specifically tailored for LTSV formatted log data.
func NewLTSVParser(writer io.Writer) *LTSVParser {
	return &LTSVParser{
		writer:      writer,
		lineDecoder: ltsvLineDecoder,
		lineHandler: JSONLineHandler,
	}
}

// SetLineHandler sets the function responsible for processing each line of log data.
// This allows for custom processing of LTSV formatted lines. The handler is skipped if nil.
func (p *LTSVParser) SetLineHandler(handler LineHandler) {
	if handler == nil {
		return
	}
	p.lineHandler = handler
}

// Parse processes log data from an io.Reader, applying the configured line and metadata handlers.
// This method supports context cancellation, prefixing of lines, and exclusion of specific lines.
func (p *LTSVParser) Parse(ctx context.Context, reader io.Reader, keywords, labels []string, hasPrefix, disableUnmatch bool) (*Result, error) {
	return parse(ctx, reader, p.writer, nil, keywords, labels, hasPrefix, disableUnmatch, p.lineDecoder, p.lineHandler)
}

// ParseString processes a log string directly, applying configured skip lines and line number handling.
// It's designed for quick parsing of a single LTSV formatted log string.
func (p *LTSVParser) ParseString(s string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error) {
	return parseString(s, p.writer, nil, keywords, labels, skipLines, hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseFile reads and parses log data from a file, leveraging the configured patterns and handlers.
// This method simplifies file-based LTSV log parsing with automatic line and metadata processing.
func (p *LTSVParser) ParseFile(filePath string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error) {
	return parseFile(filePath, p.writer, nil, keywords, labels, skipLines, hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseGzip processes gzip-compressed log data, extending the parser's capabilities to compressed LTSV logs.
// It applies skip lines and line number handling as configured for gzip-compressed files.
func (p *LTSVParser) ParseGzip(gzipPath string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error) {
	return parseGzip(gzipPath, p.writer, nil, keywords, labels, skipLines, hasLineNumber, p.lineDecoder, p.lineHandler)
}

// ParseZipEntries processes log data within zip archive entries, applying skip lines, line number handling,
// and optional glob pattern matching. This method is ideal for batch processing of LTSV logs in zip files.
func (p *LTSVParser) ParseZipEntries(zipPath, globPattern string, keywords, labels []string, skipLines []int, hasLineNumber bool) (*Result, error) {
	return parseZipEntries(zipPath, globPattern, p.writer, nil, keywords, labels, skipLines, hasLineNumber, p.lineDecoder, p.lineHandler)
}
