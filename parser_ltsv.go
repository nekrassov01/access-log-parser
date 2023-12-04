package parser

import (
	"io"
)

var _ Parser = (*LTSVParser)(nil)

// LTSVParser is a struct for parsing logs in Labeled Tab-separated Values (LTSV) format.
// LTSV format is a type of delimited text file that uses a tab to separate labels from values.
// Each line is a record with a simple key-value pair structure, making it suitable for logs and other structured data.
// The parser utilizes a custom LineHandler and MetadataHandler for processing and formatting the parsed data.
type LTSVParser struct {
	parser          parser          // parser is the underlying function that drives the LTSV parsing process.
	lineHandler     LineHandler     // LineHandler is a user-defined function that processes each line of the log file.
	metadataHandler MetadataHandler // MetadataHandler is a user-defined function that processes the metadata of the parsed log.
}

// NewLTSVParser creates a new instance of LTSVParser with default handlers.
// This parser is specifically configured to parse logs in LTSV format and format the output as JSON.
// Users can override the default handlers by using SetLineHandler and SetMetadataHandler methods.
func NewLTSVParser() *LTSVParser {
	return &LTSVParser{
		parser:          ltsvParser,          // Use the internal ltsvParser function for parsing the logs.
		lineHandler:     JSONLineHandler,     // Default handler to convert log lines to JSON format.
		metadataHandler: JSONMetadataHandler, // Default handler to convert metadata to JSON format.
	}
}

// SetLineHandler sets a custom line handler for the parser.
func (p *LTSVParser) SetLineHandler(handler LineHandler) {
	if handler == nil {
		return
	}
	p.lineHandler = handler
}

// SetMetadataHandler sets a custom metadata handler for the parser.
func (p *LTSVParser) SetMetadataHandler(handler MetadataHandler) {
	if handler == nil {
		return
	}
	p.metadataHandler = handler
}

// Parse processes the given io.Reader input using the configured patterns and handlers.
func (p *LTSVParser) Parse(input io.Reader, skipLines []int, hasIndex bool) (*Result, error) {
	return parse(input, skipLines, hasIndex, p.parser, nil, p.lineHandler, p.metadataHandler)
}

// ParseString processes the given string input as a reader.
func (p *LTSVParser) ParseString(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseString(input, skipLines, hasIndex, p.parser, nil, p.lineHandler, p.metadataHandler)
}

// ParseFile processes the content of the specified file.
func (p *LTSVParser) ParseFile(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseFile(input, skipLines, hasIndex, p.parser, nil, p.lineHandler, p.metadataHandler)
}

// ParseGzip processes the content of a gzipped file.
func (p *LTSVParser) ParseGzip(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseGzip(input, skipLines, hasIndex, p.parser, nil, p.lineHandler, p.metadataHandler)
}

// ParseZipEntries processes the contents of zip file entries matching the specified glob pattern.
func (p *LTSVParser) ParseZipEntries(input string, skipLines []int, hasIndex bool, globPattern string) ([]*Result, error) {
	return parseZipEntries(input, skipLines, hasIndex, globPattern, p.parser, nil, p.lineHandler, p.metadataHandler)
}
