package parser

import (
	"fmt"
	"io"
	"regexp"
)

var _ Parser = (*RegexParser)(nil)

type RegexParser struct {
	Patterns        []*regexp.Regexp `json:"patterns"`
	LineHandler     LineHandler      `json:"-"`
	MetadataHandler MetadataHandler  `json:"-"`
	controller      controller       `json:"-"`
}

func NewRegexParser() *RegexParser {
	return &RegexParser{
		LineHandler:     DefaultLineHandler,
		MetadataHandler: DefaultMetadataHandler,
	}
}

func (p *RegexParser) SetLineHandler(handler LineHandler) {
	p.LineHandler = handler
}

func (p *RegexParser) SetMetadataHandler(handler MetadataHandler) {
	p.MetadataHandler = handler
}

func (p *RegexParser) Parse(input io.Reader, skipLines []int) (*Result, error) {
	return parse(input, skipLines, parseRegex, p.Patterns, p.LineHandler, p.MetadataHandler)
}

func (p *RegexParser) ParseString(input string, skipLines []int) (*Result, error) {
	return parseString(input, skipLines, parseRegex, p.Patterns, p.LineHandler, p.MetadataHandler)
}

func (p *RegexParser) ParseFile(input string, skipLines []int) (*Result, error) {
	return parseFile(input, skipLines, parseRegex, p.Patterns, p.LineHandler, p.MetadataHandler)
}

func (p *RegexParser) ParseGzip(input string, skipLines []int) (*Result, error) {
	return parseGzip(input, skipLines, parseRegex, p.Patterns, p.LineHandler, p.MetadataHandler)
}

func (p *RegexParser) ParseZipEntries(input string, skipLines []int, globPattern string) ([]*Result, error) {
	return parseZipEntries(input, skipLines, globPattern, parseRegex, p.Patterns, p.LineHandler, p.MetadataHandler)
}

func (p *RegexParser) AddPattern(pattern *regexp.Regexp) error {
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

func (p *RegexParser) AddPatterns(patterns []*regexp.Regexp) error {
	for _, pattern := range patterns {
		if err := p.AddPattern(pattern); err != nil {
			p.Patterns = nil
			return err
		}
	}
	return nil
}
