package parser_test

import (
	"bytes"
	"context"
	"testing"

	parser "github.com/nekrassov01/access-log-parser"
)

func Benchmark(b *testing.B) {
	b.ResetTimer()
	line := `0.0.0.0 - - [17/May/2024:10:05:03 +0000] "GET /index.html HTTP/1.1" 200 203023 "http://your-domain.com/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/32.0.1700.77 Safari/537.36"`
	buf := &bytes.Buffer{}
	p := parser.NewS3RegexParser(context.Background(), buf, parser.Option{})
	for i := 0; i < b.N; i++ {
		if _, err := p.ParseString(line); err != nil {
			b.Fatal(err)
		}
	}
}
