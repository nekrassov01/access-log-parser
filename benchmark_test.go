package parser_test

import (
	"bytes"
	"context"
	"testing"

	parser "github.com/nekrassov01/access-log-parser"
)

func Benchmark(b *testing.B) {
	b.ResetTimer()
	line := `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -`
	buf := &bytes.Buffer{}
	p := parser.NewS3RegexParser(context.Background(), buf, parser.Option{})
	for i := 0; i < b.N; i++ {
		if _, err := p.ParseString(line); err != nil {
			b.Fatal(err)
		}
	}
}
