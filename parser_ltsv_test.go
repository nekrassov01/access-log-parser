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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			wantWriter: `{"label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLTSVParser(&bytes.Buffer{})
			got, err := p.lineHandler(tt.handlerArgs.labels, tt.handlerArgs.values, tt.handlerArgs.lineNumber, tt.handlerArgs.hasLineNumber, tt.handlerArgs.isFirst)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantWriter) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.wantWriter)
			}
		})
	}
}

func TestLTSVParser_SetLineHandler(t *testing.T) {
	type fields struct {
		writer      io.Writer
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		handler LineHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "basic",
			fields: fields{
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				handler: PrettyJSONLineHandler,
			},
			want: `{
  "no": "1",
  "label1": "value1",
  "label2": "value2",
  "label3": "value3"
}`,
		},
		{
			name: "with no handler",
			fields: fields{
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				handler: nil,
			},
			want: `{"no":"1","label1":"value1","label2":"value2","label3":"value3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				writer:      tt.fields.writer,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			p.SetLineHandler(tt.args.handler)
			got, err := p.lineHandler([]string{"label1", "label2", "label3"}, []string{"value1", "value2", "value3"}, 1, true, false)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_Parse(t *testing.T) {
	type fields struct {
		writer      io.Writer
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		ctx            context.Context
		reader         io.Reader
		keywords       []string
		labels         []string
		hasPrefix      bool
		disableUnmatch bool
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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				ctx:            context.Background(),
				reader:         strings.NewReader(ltsvAllMatchInput),
				keywords:       nil,
				labels:         nil,
				hasPrefix:      false,
				disableUnmatch: false,
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
				writer:      output,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			got, err := p.Parse(tt.args.ctx, tt.args.reader, tt.args.keywords, tt.args.labels, tt.args.hasPrefix, tt.args.disableUnmatch)
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
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		s             string
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
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
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				s:             ltsvAllMatchInput,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
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
				writer:      output,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			got, err := p.ParseString(tt.args.s, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber)
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
		writer      io.Writer
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		filePath      string
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				filePath:      filepath.Join("testdata", "sample_ltsv_all_match.log"),
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
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
				writer:      output,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			got, err := p.ParseFile(tt.args.filePath, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber)
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
		writer      io.Writer
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		gzipPath      string
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				gzipPath:      filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
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
				writer:      output,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			got, err := p.ParseGzip(tt.args.gzipPath, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber)
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
		writer      io.Writer
		lineDecoder lineDecoder
		lineHandler LineHandler
	}
	type args struct {
		zipPath       string
		globPattern   string
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
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
				writer:      &bytes.Buffer{},
				lineDecoder: ltsvLineDecoder,
				lineHandler: JSONLineHandler,
			},
			args: args{
				zipPath:       filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				globPattern:   "*",
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
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
				writer:      output,
				lineDecoder: tt.fields.lineDecoder,
				lineHandler: tt.fields.lineHandler,
			}
			got, err := p.ParseZipEntries(tt.args.zipPath, tt.args.globPattern, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber)
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
