package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewLTSVParser(t *testing.T) {
	type handlerArgs struct {
		labels   []string
		values   []string
		index    int
		hasIndex bool
	}
	tests := []struct {
		name        string
		handlerArgs handlerArgs
		want        *LTSVParser
		wantData    string
	}{
		{
			name: "with index",
			handlerArgs: handlerArgs{
				values:   []string{"value1", "value2", "value3"},
				labels:   []string{"label1", "label2", "label3"},
				index:    1,
				hasIndex: true,
			},
			want: &LTSVParser{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			wantData: `{"index":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
		{
			name: "with no index",
			handlerArgs: handlerArgs{
				values:   []string{"value1", "value2", "value3"},
				labels:   []string{"label1", "label2", "label3"},
				index:    1,
				hasIndex: false,
			},
			want: &LTSVParser{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			wantData: `{"label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLTSVParser()
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

func TestLTSVParser_SetLineHandler(t *testing.T) {
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
				parser:          ltsvParser,
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
				parser:          ltsvParser,
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
			p := &LTSVParser{
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
				t.Errorf("LTSVParser.SetLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_SetMetadataHandler(t *testing.T) {
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
				parser:          ltsvParser,
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
				parser:          ltsvParser,
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
			p := &LTSVParser{
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
				t.Errorf("LTSVParser.SetMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_Parse(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
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
			name: "ltsvParser: all match",
			fields: fields{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(ltsvAllMatchInput),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.Parse(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseString(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
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
			name: "ltsvParser: all match",
			fields: fields{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				input:     ltsvAllMatchInput,
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseFile(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
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
			name: "ltsvParser: all match",
			fields: fields{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseFile(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseGzip(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
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
			name: "ltsvParser: all match",
			fields: fields{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines: nil,
				hasIndex:  true,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log.gz"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseGzip(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseGzip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseZipEntries(t *testing.T) {
	type fields struct {
		parser          parser
		lineHandler     LineHandler
		metadataHandler MetadataHandler
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
			name: "ltsvParser: all match",
			fields: fields{
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:   nil,
				hasIndex:    true,
				globPattern: "*",
			},
			want: []*Result{
				{
					Data:     ltsvAllMatchData,
					Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				parser:          tt.fields.parser,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseZipEntries(tt.args.input, tt.args.skipLines, tt.args.hasIndex, tt.args.globPattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseZipEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseZipEntries() = %v, want %v", got, tt.want)
			}
		})
	}
}
