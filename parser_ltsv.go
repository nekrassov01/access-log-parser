package parser

import (
	"context"
	"io"
)

var _ Parser = (*LTSVParser)(nil)

// LTSVParser implements the Parser interface for parsing logs in LTSV (Labeled Tab-separated Values) format.
// It allows customization of line handling for LTSV formatted data.
type LTSVParser struct {
	ctx         context.Context
	w           io.Writer
	lineDecoder lineDecoder
	opt         Option
}

// NewLTSVParser initializes a new LTSVParser with default handlers for line decoding, line handling.
// This parser is specifically tailored for LTSV formatted log data.
func NewLTSVParser(ctx context.Context, w io.Writer, opt Option) *LTSVParser {
	p := &LTSVParser{
		ctx:         ctx,
		w:           w,
		lineDecoder: ltsvLineDecoder,
		opt:         opt,
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// Parse processes log data from an io.Reader, applying the configured line handlers.
// This method supports context cancellation, prefixing of lines, and exclusion of specific lines.
func (p *LTSVParser) Parse(reader io.Reader) (*Result, error) {
	return parse(p.ctx, reader, p.w, nil, p.lineDecoder, p.opt)
}

// ParseString processes a log string directly, applying configured skip lines and line number handling.
// It's designed for quick parsing of a single LTSV formatted log string.
func (p *LTSVParser) ParseString(s string) (*Result, error) {
	return parseString(p.ctx, s, p.w, nil, p.lineDecoder, p.opt)
}

// ParseFile reads and parses log data from a file, leveraging the configured patterns and handlers.
// This method simplifies file-based LTSV log parsing with automatic line processing.
func (p *LTSVParser) ParseFile(filePath string) (*Result, error) {
	return parseFile(p.ctx, filePath, p.w, nil, p.lineDecoder, p.opt)
}

// ParseGzip processes gzip-compressed log data, extending the parser's capabilities to compressed LTSV logs.
// It applies skip lines and line number handling as configured for gzip-compressed files.
func (p *LTSVParser) ParseGzip(gzipPath string) (*Result, error) {
	return parseGzip(p.ctx, gzipPath, p.w, nil, p.lineDecoder, p.opt)
}

// ParseZipEntries processes log data within zip archive entries, applying skip lines, line number handling,
// and optional glob pattern matching. This method is ideal for batch processing of LTSV logs in zip files.
func (p *LTSVParser) ParseZipEntries(zipPath, globPattern string) (*Result, error) {
	return parseZipEntries(p.ctx, zipPath, globPattern, p.w, nil, p.lineDecoder, p.opt)
}
