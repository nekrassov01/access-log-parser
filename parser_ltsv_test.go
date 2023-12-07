package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
				decoder:         ltsvDecoder,
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
				decoder:         ltsvDecoder,
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
		decoder         decoder
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
				decoder:         ltsvDecoder,
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
				decoder:         ltsvDecoder,
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
				decoder:         tt.fields.decoder,
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
		decoder         decoder
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
				decoder:         ltsvDecoder,
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
				decoder:         ltsvDecoder,
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
				decoder:         tt.fields.decoder,
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
		decoder         decoder
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
			name: "ltsv: all match",
			fields: fields{
				decoder:         ltsvDecoder,
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
				Labels:   ltsvAllMatchLabelData,
				Values:   ltsvAllMatchValueData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				decoder:         tt.fields.decoder,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.Parse(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Data, tt.want.Data); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Labels, tt.want.Labels); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Values, tt.want.Values); diff != "" {
				t.Errorf(diff)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseString(t *testing.T) {
	type fields struct {
		decoder         decoder
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
			name: "ltsv: all match",
			fields: fields{
				decoder:         ltsvDecoder,
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
				Labels:   ltsvAllMatchLabelData,
				Values:   ltsvAllMatchValueData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				decoder:         tt.fields.decoder,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseString(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Data, tt.want.Data); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Labels, tt.want.Labels); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Values, tt.want.Values); diff != "" {
				t.Errorf(diff)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseFile(t *testing.T) {
	type fields struct {
		decoder         decoder
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
			name: "ltsv: all match",
			fields: fields{
				decoder:         ltsvDecoder,
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
				Labels:   ltsvAllMatchLabelData,
				Values:   ltsvAllMatchValueData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				decoder:         tt.fields.decoder,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseFile(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Data, tt.want.Data); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Labels, tt.want.Labels); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Values, tt.want.Values); diff != "" {
				t.Errorf(diff)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LTSVParser.ParseFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVParser_ParseGzip(t *testing.T) {
	type fields struct {
		decoder         decoder
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
			name: "ltsv: all match",
			fields: fields{
				decoder:         ltsvDecoder,
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
				Labels:   ltsvAllMatchLabelData,
				Values:   ltsvAllMatchValueData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				decoder:         tt.fields.decoder,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			got, err := p.ParseGzip(tt.args.input, tt.args.skipLines, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got.Data, tt.want.Data); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Labels, tt.want.Labels); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(got.Values, tt.want.Values); diff != "" {
				t.Errorf(diff)
			}
			if !reflect.DeepEqual(got.Data, tt.want.Data) {
				t.Errorf("LTSVParser.ParseGzip() = %v, want %v", got.Data, tt.want.Data)
			}
		})
	}
}

func TestLTSVParser_ParseZipEntries(t *testing.T) {
	type fields struct {
		decoder         decoder
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
			name: "ltsv: all match",
			fields: fields{
				decoder:         ltsvDecoder,
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
					Labels:   ltsvAllMatchLabelData,
					Values:   ltsvAllMatchValueData,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LTSVParser{
				decoder:         tt.fields.decoder,
				lineHandler:     tt.fields.lineHandler,
				metadataHandler: tt.fields.metadataHandler,
			}
			gots, err := p.ParseZipEntries(tt.args.input, tt.args.skipLines, tt.args.hasIndex, tt.args.globPattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVParser.ParseZipEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, got := range gots {
				for _, w := range tt.want {
					if diff := cmp.Diff(got.Data, w.Data); diff != "" {
						t.Errorf(diff)
					}
				}
			}
			for _, got := range gots {
				for _, w := range tt.want {
					if diff := cmp.Diff(got.Labels, w.Labels); diff != "" {
						t.Errorf(diff)
					}
				}
			}
			for _, got := range gots {
				for _, w := range tt.want {
					if diff := cmp.Diff(got.Values, w.Values); diff != "" {
						t.Errorf(diff)
					}
				}
			}
			if !reflect.DeepEqual(gots, tt.want) {
				t.Errorf("LTSVParser.ParseZipEntries() = %v, want %v", gots, tt.want)
			}
		})
	}
}
