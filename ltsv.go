package parser

import (
	"io"
)

var _ Parser = (*LTSVParser)(nil)

// LTSVParser is a struct that defines how to parse the access log.
// Patterns represents a list of regular expression patterns used for matching log lines,
// which are matched in order. Each field must have a named capture group.
// LineHandler is a custom function that processes the matched log lines.
// If not provided at instantiation, DefaultLineHandler will be used.
// MetadataHandler is a custom function that processes and formats metadata.
// If not provided at instantiation, DefaultMetadataHandler will be used.
type LTSVParser struct {
	LineHandler     LineHandler     `json:"-"`
	MetadataHandler MetadataHandler `json:"-"`
	controller      controller      `json:"-"`
}

// NewLTSVParser creates a new parser with the specified options.
func NewLTSVParser() *LTSVParser {
	return &LTSVParser{
		LineHandler:     DefaultLineHandler,
		MetadataHandler: DefaultMetadataHandler,
		controller:      parseLTSV,
	}
}

func (p *LTSVParser) SetLineHandler(handler LineHandler) {
	p.LineHandler = handler
}

func (p *LTSVParser) SetMetadataHandler(handler MetadataHandler) {
	p.MetadataHandler = handler
}

// Parse parses the provided io.Reader and returns a Result.
// Use this if you want to use an abstracted Reader.
func (p *LTSVParser) Parse(input io.Reader, skipLines []int) (*Result, error) {
	return parse(input, skipLines, parseLTSV, nil, p.LineHandler, p.MetadataHandler)
}

// ParseString parses the provided string input and returns a Result.
func (p *LTSVParser) ParseString(input string, skipLines []int) (*Result, error) {
	return parseString(input, skipLines, parseLTSV, nil, p.LineHandler, p.MetadataHandler)
}

// ParseFile parsers the content of a file and returns a Result.
func (p *LTSVParser) ParseFile(input string, skipLines []int) (*Result, error) {
	return parseFile(input, skipLines, parseLTSV, nil, p.LineHandler, p.MetadataHandler)
}

// ParseGzip parsers the content of a gzipped file and returns a Result.
func (p *LTSVParser) ParseGzip(input string, skipLines []int) (*Result, error) {
	return parseGzip(input, skipLines, parseLTSV, nil, p.LineHandler, p.MetadataHandler)
}

// ParseZipEntries parses the contents of entries in the zip archive that match the glob pattern and returns the result.
func (p *LTSVParser) ParseZipEntries(input string, skipLines []int, globPattern string) ([]*Result, error) {
	return parseZipEntries(input, skipLines, globPattern, parseLTSV, nil, p.LineHandler, p.MetadataHandler)
}
