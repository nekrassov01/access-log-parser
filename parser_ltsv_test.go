package parser

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewLTSVParser(t *testing.T) {
	type handlerArgs struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name        string
		handlerArgs handlerArgs
		want        *LTSVParser
		wantWriter  string
	}{
		{
			name: "with lineNumber",
			handlerArgs: handlerArgs{
				labels:        []string{"label1", "label2", "label3"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    1,
				hasLineNumber: true,
			}, want: &LTSVParser{
				writer:  &bytes.Buffer{},
				decoder: ltsvLineDecoder,
			},
			wantWriter: `{"no":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
		{
			name: "with no lineNumber",
			handlerArgs: handlerArgs{
				labels:        []string{"label1", "label2", "label3"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    1,
				hasLineNumber: false,
			}, want: &LTSVParser{
				writer:  &bytes.Buffer{},
				decoder: ltsvLineDecoder,
			},
			wantWriter: `{"label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLTSVParser(context.Background(), &bytes.Buffer{}, Option{})
			got, err := p.opt.LineHandler(tt.handlerArgs.labels, tt.handlerArgs.values, tt.handlerArgs.lineNumber, tt.handlerArgs.hasLineNumber, tt.handlerArgs.isFirst)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantWriter) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.wantWriter)
			}
		})
	}
}

func TestLTSVParser_Parse(t *testing.T) {
	type fields struct {
		ctx     context.Context
		decoder lineDecoder
		opt     Option
	}
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "ltsv: all match",
			fields: fields{
				ctx:     context.Background(),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
			},
			args: args{
				reader: strings.NewReader(ltsvAllMatchInput),
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "",
				inputType: inputTypeStream,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			p := &LTSVParser{
				ctx:     tt.fields.ctx,
				writer:  output,
				decoder: ltsvLineDecoder,
				opt:     tt.fields.opt,
			}
			got, err := p.Parse(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func TestLTSVParser_ParseString(t *testing.T) {
	type fields struct {
		ctx     context.Context
		decoder lineDecoder
		opt     Option
	}
	type args struct {
		s string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "ltsv: all match",
			fields: fields{
				ctx:     context.Background(),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
			},
			args: args{
				s: ltsvAllMatchInput,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "",
				inputType: inputTypeString,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			p := &LTSVParser{
				ctx:     tt.fields.ctx,
				writer:  output,
				decoder: tt.fields.decoder,
				opt:     tt.fields.opt,
			}
			got, err := p.ParseString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func TestLTSVParser_ParseFile(t *testing.T) {
	type fields struct {
		ctx     context.Context
		decoder lineDecoder
		opt     Option
	}
	type args struct {
		filePath string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "ltsv: all match",
			fields: fields{
				ctx:     context.Background(),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
			},
			args: args{
				filePath: filepath.Join("testdata", "sample_ltsv_all_match.log"),
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_ltsv_all_match.log",
				inputType: inputTypeFile,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			p := &LTSVParser{
				ctx:     tt.fields.ctx,
				writer:  output,
				decoder: tt.fields.decoder,
				opt:     tt.fields.opt,
			}
			got, err := p.ParseFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func TestLTSVParser_ParseGzip(t *testing.T) {
	type fields struct {
		ctx     context.Context
		decoder lineDecoder
		opt     Option
	}
	type args struct {
		gzipPath string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "ltsv: all match",
			fields: fields{
				ctx:     context.Background(),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
			},
			args: args{
				gzipPath: filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_ltsv_all_match.log.gz",
				inputType: inputTypeGzip,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			p := &LTSVParser{
				ctx:     tt.fields.ctx,
				writer:  output,
				decoder: tt.fields.decoder,
				opt:     tt.fields.opt,
			}
			got, err := p.ParseGzip(tt.args.gzipPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func TestLTSVParser_ParseZipEntries(t *testing.T) {
	type fields struct {
		ctx     context.Context
		decoder lineDecoder
		opt     Option
	}
	type args struct {
		zipPath     string
		globPattern string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "ltsv: all match",
			fields: fields{
				ctx:     context.Background(),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
			},
			args: args{
				zipPath:     filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				globPattern: "*",
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_ltsv_all_match.log.zip",
				inputType: inputTypeZip,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			p := &LTSVParser{
				ctx:     tt.fields.ctx,
				writer:  output,
				decoder: tt.fields.decoder,
				opt:     tt.fields.opt,
			}
			got, err := p.ParseZipEntries(tt.args.zipPath, tt.args.globPattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}
