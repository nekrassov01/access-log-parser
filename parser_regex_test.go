package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestNewRegexParser(t *testing.T) {
	type handlerArgs struct {
		labels   []string
		values   []string
		index    int
		hasIndex bool
	}
	tests := []struct {
		name        string
		handlerArgs handlerArgs
		want        *RegexParser
		wantData    string
	}{
		{
			name: "with index",
			handlerArgs: handlerArgs{
				labels:   []string{"label1", "label2", "label3"},
				values:   []string{"value1", "value2", "value3"},
				index:    1,
				hasIndex: true,
			},
			want: &RegexParser{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			wantData: `{"index":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
		{
			name: "with no index",
			handlerArgs: handlerArgs{
				labels:   []string{"label1", "label2", "label3"},
				values:   []string{"value1", "value2", "value3"},
				index:    1,
				hasIndex: false,
			},
			want: &RegexParser{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			wantData: `{"label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewRegexParser()
			got1, err := p.lineHandler(tt.handlerArgs.labels, tt.handlerArgs.values, tt.handlerArgs.index, tt.handlerArgs.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got1, tt.wantData) {
				t.Errorf("NewRegexParser() = %v, want %v", got1, tt.wantData)
			}
			metadata := regexAllMatchMetadata
			wantMetadata := fmt.Sprintf(regexAllMatchMetadataSerialized, "")
			got2, err := p.metadataHandler(metadata)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got2, wantMetadata) {
				t.Errorf("NewRegexParser() = %v, want %v", got1, wantMetadata)
			}
		})
	}
}

func TestRegexParser_SetLineHandler(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	type args struct {
		handler LineHandler
	}
	type handlerArgs struct {
		labels   []string
		values   []string
		index    int
		hasIndex bool
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		handlerArgs handlerArgs
		want        string
	}{
		{
			name: "basic",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				handler: JSONLineHandler,
			},
			handlerArgs: handlerArgs{
				labels:   []string{"label1", "label2", "label3"},
				values:   []string{"value1", "value2", "value3"},
				index:    1,
				hasIndex: true,
			},
			want: `{"index":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
		{
			name: "nil input",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				handler: nil,
			},
			handlerArgs: handlerArgs{
				labels:   []string{"label1", "label2", "label3"},
				values:   []string{"value1", "value2", "value3"},
				index:    1,
				hasIndex: true,
			},
			want: `{"index":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			p.SetLineHandler(tt.args.handler)
			got, err := p.lineHandler(tt.handlerArgs.labels, tt.handlerArgs.values, tt.handlerArgs.index, tt.handlerArgs.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.SetLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_SetMetadataHandler(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	type args struct {
		handler MetadataHandler
	}
	type handlerArgs struct {
		metadata *Metadata
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		handlerArgs handlerArgs
		want        string
	}{
		{
			name: "basic",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				handler: JSONMetadataHandler,
			},
			handlerArgs: handlerArgs{
				metadata: &Metadata{
					Total:     5,
					Matched:   5,
					Unmatched: 0,
					Skipped:   0,
					Source:    "",
					Errors:    nil,
				},
			},
			want: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
		},
		{
			name: "nil input",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				handler: nil,
			},
			handlerArgs: handlerArgs{
				metadata: &Metadata{
					Total:     5,
					Matched:   5,
					Unmatched: 0,
					Skipped:   0,
					Source:    "",
					Errors:    nil,
				},
			},
			want: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			p.SetMetadataHandler(tt.args.handler)
			got, err := p.metadataHandler(tt.handlerArgs.metadata)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.SetMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_Parse(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		input     io.Reader
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			got, err := p.Parse(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_ParseString(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			args: args{
				input:     regexAllMatchInput,
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.ParseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_ParseFile(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			got, err := p.ParseFile(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.ParseFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_ParseGzip(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log.gz"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			got, err := p.ParseGzip(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.ParseGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.ParseGzip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_ParseZipEntries(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		input       string
		skipLines   []int
		hasIndex    bool
		globPattern string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				hasIndex:    true,
				globPattern: "*",
			},
			want: []*Result{
				{
					Data:     regexAllMatchData,
					Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			got, err := p.ParseZipEntries(tt.args.input, tt.args.skipLines, tt.args.hasIndex, tt.args.globPattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.ParseZipEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.ParseZipEntries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexParser_Decode(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		want1   []string
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns: []*regexp.Regexp{
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
				},
			},
			args: args{
				s: `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
			},
			want:    []string{"remote_host", "remote_logname", "remote_user", "datetime", "method", "request_uri", "protocol", "status", "size", "referer", "user_agent"},
			want1:   []string{"123.45.67.89", "-", "frank", "[10/Oct/2000:13:55:36 -0700]", "GET", "/apache_pb.gif", "HTTP/1.0", "200", "2326", "http://www.example.com/start.html", "Mozilla/4.08 [en] (Win98; I ;Nav)"},
			wantErr: false,
		},
		{
			name: "invalid input",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns: []*regexp.Regexp{
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
				},
			},
			args: args{
				s: "aaa",
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "blank input",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns: []*regexp.Regexp{
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
					regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
				},
			},
			args: args{
				s: "",
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "nil pattern",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        nil,
			},
			args: args{
				s: `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			labels, values, err := p.Decode(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(labels, tt.want) {
				t.Errorf("RegexParser.Decode() = %v, want %v", labels, tt.want)
			}
			if !reflect.DeepEqual(values, tt.want1) {
				t.Errorf("RegexParser.Decode() = %v, want %v", values, tt.want1)
			}
		})
	}
}

func TestRegexParser_AddPattern(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		pattern *regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RegexParser
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				pattern: regexPattern,
			},
			want:    &RegexParser{patterns: []*regexp.Regexp{regexPattern}},
			wantErr: false,
		},
		{
			name: "caputure group not found",
			args: args{
				pattern: regexCapturedGroupNotContainsPattern,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "non-named caputure group detected",
			args: args{
				pattern: regexNonNamedCapturedGroupContainsPattern,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			if err := p.AddPattern(tt.args.pattern); (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.AddPattern() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegexParser_AddPatterns(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	type args struct {
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RegexParser
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				patterns: regexPatterns,
			},
			want: &RegexParser{
				patterns: regexPatterns,
			},
			wantErr: false,
		},
		{
			name: "caputure group not found",
			args: args{
				patterns: regexCapturedGroupNotContainsPatterns,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "non-named caputure group detected",
			args: args{
				patterns: regexNonNamedCapturedGroupContainsPatterns,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			if err := p.AddPatterns(tt.args.patterns); (err != nil) != tt.wantErr {
				t.Errorf("RegexParser.AddPatterns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegexParser_GetPatterns(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
		patterns        []*regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*regexp.Regexp
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				parser:          regexParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
				patterns:        regexPatterns,
			},
			want:    regexPatterns,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RegexParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
				patterns:        tt.fields.patterns,
			}
			if got := p.GetPatterns(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexParser.GetPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewApacheCLFRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "combined",
			args: args{
				input:     `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "combined+noindex",
			args: args{
				input:     `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "common",
			args: args{
				input:     `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326"}`},
			wantErr: false,
		},
		{
			name: "combined+tab",
			args: args{
				input:     `123.45.67.89	-	frank	[10/Oct/2000:13:55:36 -0700]	"GET /apache_pb.gif HTTP/1.0"	200	2326	"http://www.example.com/start.html"	"Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "common+tab",
			args: args{
				input:     `123.45.67.89	-	frank	[10/Oct/2000:13:55:36 -0700]	"GET /apache_pb.gif HTTP/1.0"	200	2326`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326"}`},
			wantErr: false,
		},
		{
			name: "remote_user contains space",
			args: args{
				input:     `123.45.67.89 - frank zappa [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank zappa","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewApacheCLFRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewApacheCLFRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewApacheCLFWithVHostRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "combined",
			args: args{
				input:     `example.com 123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "combined+noindex",
			args: args{
				input:     `example.com 123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "common",
			args: args{
				input:     `example.com 123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326"}`},
			wantErr: false,
		},
		{
			name: "combined+tab",
			args: args{
				input:     `example.com	123.45.67.89	-	frank	[10/Oct/2000:13:55:36 -0700]	"GET /apache_pb.gif HTTP/1.0"	200	2326	"http://www.example.com/start.html"	"Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "common+tab",
			args: args{
				input:     `example.com	123.45.67.89	-	frank	[10/Oct/2000:13:55:36 -0700]	"GET /apache_pb.gif HTTP/1.0"	200	2326`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326"}`},
			wantErr: false,
		},
		{
			name: "remote_user contains space",
			args: args{
				input:     `example.com 123.45.67.89 - frank zappa [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","virtual_host":"example.com","remote_host":"123.45.67.89","remote_logname":"-","remote_user":"frank zappa","datetime":"[10/Oct/2000:13:55:36 -0700]","method":"GET","request_uri":"/apache_pb.gif","protocol":"HTTP/1.0","status":"200","size":"2326","referer":"http://www.example.com/start.html","user_agent":"Mozilla/4.08 [en] (Win98; I ;Nav)"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `example.com 123.45.67.89 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewApacheCLFWithVHostRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewApacheCLFWithVHostRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewS3RegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				input:     `01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`},
			wantErr: false,
		},
		{
			name: "1st+noindex",
			args: args{
				input:     `01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`},
			wantErr: false,
		},
		{
			name: "2nd",
			args: args{
				input:     `d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket89?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`},
			wantErr: false,
		},
		{
			name: "3rd",
			args: args{
				input:     `8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`},
			wantErr: false,
		},
		{
			name: "4th",
			args: args{
				input:     `3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`},
			wantErr: false,
		},
		{
			name: "5th",
			args: args{
				input:     `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewS3RegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewS3RegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewCFRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				input:     `2019-12-04	21:02:31	LAX1	392	192.0.2.100	GET	d111111abcdef8.cloudfront.net	/index.html	200	-	Mozilla/5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/78.0.3904.108%20Safari/537.36	-	-	Hit	SOX4xwn4XV6Q4rgb7XiVGOHms_BGlTAC4KyHmureZmBNrjGdRLiNIQ==	d111111abcdef8.cloudfront.net	https	23	0.001	-	TLSv1.2	ECDHE-RSA-AES128-GCM-SHA256	Hit	HTTP/2.0	-	-	11040	0.001	Hit	text/html	78	-	-`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","date":"2019-12-04","time":"21:02:31","x_edge_location":"LAX1","sc_bytes":"392","c_ip":"192.0.2.100","cs_method":"GET","cs_host":"d111111abcdef8.cloudfront.net","cs_uri_stem":"/index.html","sc_status":"200","cs_referer":"-","cs_user_agent":"Mozilla/5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/78.0.3904.108%20Safari/537.36","cs_uri_query":"-","cs_cookie":"-","x_edge_result_type":"Hit","x_edge_request_id":"SOX4xwn4XV6Q4rgb7XiVGOHms_BGlTAC4KyHmureZmBNrjGdRLiNIQ==","x_host_header":"d111111abcdef8.cloudfront.net","cs_protocol":"https","cs_bytes":"23","time_taken":"0.001","x_forwarded_for":"-","ssl_protocol":"TLSv1.2","ssl_cipher":"ECDHE-RSA-AES128-GCM-SHA256","x_edge_response_result_type":"Hit","cs_protocol_version":"HTTP/2.0","fle_status":"-","fle_encrypted_fields":"-","c_port":"11040","time_to_first_byte":"0.001","x_edge_detailed_result_type":"Hit","sc_content_type":"text/html","sc_content_len":"78","sc_range_start":"-","sc_range_end":"-"}`},
			wantErr: false,
		},
		{
			name: "1st+noindex",
			args: args{
				input:     `2019-12-04	21:02:31	LAX1	392	192.0.2.100	GET	d111111abcdef8.cloudfront.net	/index.html	200	-	Mozilla/5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/78.0.3904.108%20Safari/537.36	-	-	Hit	SOX4xwn4XV6Q4rgb7XiVGOHms_BGlTAC4KyHmureZmBNrjGdRLiNIQ==	d111111abcdef8.cloudfront.net	https	23	0.001	-	TLSv1.2	ECDHE-RSA-AES128-GCM-SHA256	Hit	HTTP/2.0	-	-	11040	0.001	Hit	text/html	78	-	-`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"date":"2019-12-04","time":"21:02:31","x_edge_location":"LAX1","sc_bytes":"392","c_ip":"192.0.2.100","cs_method":"GET","cs_host":"d111111abcdef8.cloudfront.net","cs_uri_stem":"/index.html","sc_status":"200","cs_referer":"-","cs_user_agent":"Mozilla/5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/78.0.3904.108%20Safari/537.36","cs_uri_query":"-","cs_cookie":"-","x_edge_result_type":"Hit","x_edge_request_id":"SOX4xwn4XV6Q4rgb7XiVGOHms_BGlTAC4KyHmureZmBNrjGdRLiNIQ==","x_host_header":"d111111abcdef8.cloudfront.net","cs_protocol":"https","cs_bytes":"23","time_taken":"0.001","x_forwarded_for":"-","ssl_protocol":"TLSv1.2","ssl_cipher":"ECDHE-RSA-AES128-GCM-SHA256","x_edge_response_result_type":"Hit","cs_protocol_version":"HTTP/2.0","fle_status":"-","fle_encrypted_fields":"-","c_port":"11040","time_to_first_byte":"0.001","x_edge_detailed_result_type":"Hit","sc_content_type":"text/html","sc_content_len":"78","sc_range_start":"-","sc_range_end":"-"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `2019-12-04	21:02:31	LAX1	392	192.0.2.100	GET	d111111abcdef8.cloudfront.net	/index.html	200	-	Mozilla/5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/78.0.3904.108%20Safari/537.36	-	-	Hit	SOX4xwn4XV6Q4rgb7XiVGOHms_BGlTAC4KyHmureZmBNrjGdRLiNIQ==	d111111abcdef8.cloudfront.net	https	23	0.001	-	TLSv1.2	ECDHE-RSA-AES128-GCM-SHA256	Hit	HTTP/2.0	-	-	11040	0.001	Hit	text/html	78	-`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewCFRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewCFRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewALBRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				input:     `http 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.000 0.001 0.000 200 200 34 366 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.46.0" - - arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337262-36d228ad5d99923122bbe354" "-" "-" 0 2018-07-02T22:22:48.364000Z "forward" "-" "-" "10.0.0.1:80" "200" "-" "-"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","type":"http","time":"2018-07-02T22:23:00.186641Z","elb":"app/my-loadbalancer/50dc6c495c0c9188","client_port":"192.168.131.39:2817","target_port":"10.0.0.1:80","request_processing_time":"0.000","target_processing_time":"0.001","response_processing_time":"0.000","elb_status_code":"200","target_status_code":"200","received_bytes":"34","sent_bytes":"366","method":"GET","request_uri":"http://www.example.com:80/","protocol":"HTTP/1.1","user_agent":"curl/7.46.0","ssl_cipher":"-","ssl_protocol":"-","target_group_arn":"arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067","trace_id":"Root=1-58337262-36d228ad5d99923122bbe354","domain_name":"-","chosen_cert_arn":"-","matched_rule_priority":"0","request_creation_time":"2018-07-02T22:22:48.364000Z","actions_executed":"forward","redirect_url":"-","error_reason":"-","target_port_list":"10.0.0.1:80","target_status_code_list":"200","classification":"-","classification_reason":"-"}`},
			wantErr: false,
		},
		{
			name: "1st+noindex",
			args: args{
				input:     `http 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.000 0.001 0.000 200 200 34 366 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.46.0" - - arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337262-36d228ad5d99923122bbe354" "-" "-" 0 2018-07-02T22:22:48.364000Z "forward" "-" "-" "10.0.0.1:80" "200" "-" "-"`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"type":"http","time":"2018-07-02T22:23:00.186641Z","elb":"app/my-loadbalancer/50dc6c495c0c9188","client_port":"192.168.131.39:2817","target_port":"10.0.0.1:80","request_processing_time":"0.000","target_processing_time":"0.001","response_processing_time":"0.000","elb_status_code":"200","target_status_code":"200","received_bytes":"34","sent_bytes":"366","method":"GET","request_uri":"http://www.example.com:80/","protocol":"HTTP/1.1","user_agent":"curl/7.46.0","ssl_cipher":"-","ssl_protocol":"-","target_group_arn":"arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067","trace_id":"Root=1-58337262-36d228ad5d99923122bbe354","domain_name":"-","chosen_cert_arn":"-","matched_rule_priority":"0","request_creation_time":"2018-07-02T22:22:48.364000Z","actions_executed":"forward","redirect_url":"-","error_reason":"-","target_port_list":"10.0.0.1:80","target_status_code_list":"200","classification":"-","classification_reason":"-"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `http 2018-07-02T22:23:00.186641Z app/my-loadbalancer/50dc6c495c0c9188 192.168.131.39:2817 10.0.0.1:80 0.000 0.001 0.000 200 200 34 366 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.46.0" - - arn:aws:elasticloadbalancing:us-east-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067 "Root=1-58337262-36d228ad5d99923122bbe354" "-" "-" 0 2018-07-02T22:22:48.364000Z "forward" "-" "-" 10.0.0.1:80 200 "-"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewALBRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewALBRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewNLBRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				input:     `tls 2.0 2018-12-20T02:59:40 net/my-network-loadbalancer/c6e77e28c25b2234 g3d4b5e8bb8464cd 72.21.218.154:51341 172.100.100.185:443 5 2 98 246 - arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99 - ECDHE-RSA-AES128-SHA tlsv12 - my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com - - - 2018-12-20T02:59:30`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","type":"tls","version":"2.0","time":"2018-12-20T02:59:40","elb":"net/my-network-loadbalancer/c6e77e28c25b2234","listener":"g3d4b5e8bb8464cd","client_port":"72.21.218.154:51341","destination_port":"172.100.100.185:443","connection_time":"5","tls_handshake_time":"2","received_bytes":"98","sent_bytes":"246","incoming_tls_alert":"-","chosen_cert_arn":"arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99","chosen_cert_serial":"-","tls_cipher":"ECDHE-RSA-AES128-SHA","tls_protocol_version":"tlsv12","tls_named_group":"-","domain_name":"my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com","alpn_fe_protocol":"-","alpn_be_protocol":"-","alpn_client_preference_list":"-","tls_connection_creation_time":"2018-12-20T02:59:30"}`},
			wantErr: false,
		},
		{
			name: "1st+noindex",
			args: args{
				input:     `tls 2.0 2018-12-20T02:59:40 net/my-network-loadbalancer/c6e77e28c25b2234 g3d4b5e8bb8464cd 72.21.218.154:51341 172.100.100.185:443 5 2 98 246 - arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99 - ECDHE-RSA-AES128-SHA tlsv12 - my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com - - - 2018-12-20T02:59:30`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"type":"tls","version":"2.0","time":"2018-12-20T02:59:40","elb":"net/my-network-loadbalancer/c6e77e28c25b2234","listener":"g3d4b5e8bb8464cd","client_port":"72.21.218.154:51341","destination_port":"172.100.100.185:443","connection_time":"5","tls_handshake_time":"2","received_bytes":"98","sent_bytes":"246","incoming_tls_alert":"-","chosen_cert_arn":"arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99","chosen_cert_serial":"-","tls_cipher":"ECDHE-RSA-AES128-SHA","tls_protocol_version":"tlsv12","tls_named_group":"-","domain_name":"my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com","alpn_fe_protocol":"-","alpn_be_protocol":"-","alpn_client_preference_list":"-","tls_connection_creation_time":"2018-12-20T02:59:30"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `tls 2.0 2018-12-20T02:59:40 net/my-network-loadbalancer/c6e77e28c25b2234 g3d4b5e8bb8464cd 72.21.218.154:51341 172.100.100.185:443 5 2 98 246 - arn:aws:acm:us-east-2:671290407336:certificate/2a108f19-aded-46b0-8493-c63eb1ef4a99 - ECDHE-RSA-AES128-SHA tlsv12 - my-network-loadbalancer-c6e77e28c25b2234.elb.us-east-2.amazonaws.com - - -`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewNLBRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewNLBRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestNewCLBRegexParser(t *testing.T) {
	type args struct {
		input     string
		skipLines []int
		hasIndex  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				input:     `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","time":"2015-05-13T23:39:43.945958Z","elb":"my-loadbalancer","client_port":"192.168.131.39:2817","backend_port":"10.0.0.1:80","request_processing_time":"0.000073","backend_processing_time":"0.001048","response_processing_time":"0.000057","elb_status_code":"200","backend_status_code":"200","received_bytes":"0","sent_bytes":"29","method":"GET","request_uri":"http://www.example.com:80/","protocol":"HTTP/1.1","user_agent":"curl/7.38.0","ssl_cipher":"-","ssl_protocol":"-"}`},
			wantErr: false,
		},
		{
			name: "1st+noindex",
			args: args{
				input:     `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`,
				skipLines: nil,
				hasIndex:  false,
			},
			want:    []string{`{"time":"2015-05-13T23:39:43.945958Z","elb":"my-loadbalancer","client_port":"192.168.131.39:2817","backend_port":"10.0.0.1:80","request_processing_time":"0.000073","backend_processing_time":"0.001048","response_processing_time":"0.000057","elb_status_code":"200","backend_status_code":"200","received_bytes":"0","sent_bytes":"29","method":"GET","request_uri":"http://www.example.com:80/","protocol":"HTTP/1.1","user_agent":"curl/7.38.0","ssl_cipher":"-","ssl_protocol":"-"}`},
			wantErr: false,
		},
		{
			name: "2st",
			args: args{
				input:     `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1"`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    []string{`{"index":"1","time":"2015-05-13T23:39:43.945958Z","elb":"my-loadbalancer","client_port":"192.168.131.39:2817","backend_port":"10.0.0.1:80","request_processing_time":"0.000073","backend_processing_time":"0.001048","response_processing_time":"0.000057","elb_status_code":"200","backend_status_code":"200","received_bytes":"0","sent_bytes":"29","method":"GET","request_uri":"http://www.example.com:80/","protocol":"HTTP/1.1"}`},
			wantErr: false,
		},
		{
			name: "unmatch",
			args: args{
				input:     `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29`,
				skipLines: nil,
				hasIndex:  true,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewCLBRegexParser()
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("NewCLBRegexParser() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}
