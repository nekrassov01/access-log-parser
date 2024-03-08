package parser

import (
	"reflect"
	"testing"
)

func TestJSONLineHandler(t *testing.T) {
	type args struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2"},
				lineNumber:    1,
				hasLineNumber: true,
			},
			want:    `{"no":"1","label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"\"value1", "value2"},
				lineNumber:    2,
				hasLineNumber: true,
			},
			want:    `{"no":"2","label1":"\"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    3,
				hasLineNumber: true,
			},
			want:    `{"no":"3","label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1"},
				lineNumber:    4,
				hasLineNumber: true,
			},
			want:    `{"no":"4","label1":"value1"}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value 2"},
				lineNumber:    5,
				hasLineNumber: true,
			},
			want:    `{"no":"5","label1":"value1","label2":"value 2"}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"-", "\"-\""},
				lineNumber:    6,
				hasLineNumber: true,
			},
			want:    `{"no":"6","label1":"-","label2":"\"-\""}`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", `value\2`},
				lineNumber:    7,
				hasLineNumber: true,
			},
			want:    `{"no":"7","label1":"value1","label2":"value\\2"}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONLineHandler(tt.args.labels, tt.args.values, tt.args.lineNumber, tt.args.hasLineNumber, tt.args.isFirst)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func TestPrettyJSONLineHandler(t *testing.T) {
	type args struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2"},
				lineNumber:    1,
				hasLineNumber: true,
			},
			want: `{
  "no": "1",
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"\"value1", "value2"},
				lineNumber:    2,
				hasLineNumber: true,
			},
			want: `{
  "no": "2",
  "label1": "\"value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    3,
				hasLineNumber: true,
			},
			want: `{
  "no": "3",
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1"},
				lineNumber:    4,
				hasLineNumber: true,
			},
			want: `{
  "no": "4",
  "label1": "value1"
}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value 2"},
				lineNumber:    5,
				hasLineNumber: true,
			},
			want: `{
  "no": "5",
  "label1": "value1",
  "label2": "value 2"
}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"-", "\"-\""},
				lineNumber:    6,
				hasLineNumber: true,
			},
			want: `{
  "no": "6",
  "label1": "-",
  "label2": "\"-\""
}`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", `value\2`},
				lineNumber:    7,
				hasLineNumber: true,
			},
			want: `{
  "no": "7",
  "label1": "value1",
  "label2": "value\\2"
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSONLineHandler(tt.args.labels, tt.args.values, tt.args.lineNumber, tt.args.hasLineNumber, tt.args.isFirst)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func TestKeyValuePairLineHandler(t *testing.T) {
	type args struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2"},
				lineNumber:    1,
				hasLineNumber: true,
			},
			want:    `no="1" label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"\"value1", "value2"},
				lineNumber:    2,
				hasLineNumber: true,
			},
			want:    `no="2" label1="\"value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    3,
				hasLineNumber: true,
			},
			want:    `no="3" label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1"},
				lineNumber:    4,
				hasLineNumber: true,
			},
			want:    `no="4" label1="value1"`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value 2"},
				lineNumber:    5,
				hasLineNumber: true,
			},
			want:    `no="5" label1="value1" label2="value 2"`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"-", "\"-\""},
				lineNumber:    6,
				hasLineNumber: true,
			},
			want:    `no="6" label1="-" label2="\"-\""`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", `value\2`},
				lineNumber:    7,
				hasLineNumber: true,
			},
			want:    `no="7" label1="value1" label2="value\\2"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyValuePairLineHandler(tt.args.labels, tt.args.values, tt.args.lineNumber, tt.args.hasLineNumber, tt.args.isFirst)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func TestLTSVLineHandler(t *testing.T) {
	type args struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2"},
				lineNumber:    1,
				hasLineNumber: true,
			},
			want:    `no:1	label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"\"value1", "value2"},
				lineNumber:    2,
				hasLineNumber: true,
			},
			want:    `no:2	label1:"value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    3,
				hasLineNumber: true,
			},
			want:    `no:3	label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1"},
				lineNumber:    4,
				hasLineNumber: true,
			},
			want:    `no:4	label1:value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value 2"},
				lineNumber:    5,
				hasLineNumber: true,
			},
			want:    `no:5	label1:value1	label2:value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"-", "\"-\""},
				lineNumber:    6,
				hasLineNumber: true,
			},
			want:    `no:6	label1:-	label2:"-"`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", `value\2`},
				lineNumber:    7,
				hasLineNumber: true,
			},
			want:    `no:7	label1:value1	label2:value\2`,
			wantErr: false,
		},
		{
			name: "empty string included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"", `value2`},
				lineNumber:    7,
				hasLineNumber: true,
			},
			want:    `no:7	label1:-	label2:value2`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LTSVLineHandler(tt.args.labels, tt.args.values, tt.args.lineNumber, tt.args.hasLineNumber, tt.args.isFirst)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func TestTSVLineHandler(t *testing.T) {
	type args struct {
		labels        []string
		values        []string
		lineNumber    int
		hasLineNumber bool
		isFirst       bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "123"},
				lineNumber:    1,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `1	value1	123`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"\"value1", "123"},
				lineNumber:    2,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `2	"value1	123`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value2", "value3"},
				lineNumber:    3,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `3	value1	value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1"},
				lineNumber:    4,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `4	value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "value 2"},
				lineNumber:    5,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `5	value1	value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"-", "\"-\""},
				lineNumber:    6,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `6	-	"-"`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", `value\2`},
				lineNumber:    7,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `7	value1	value\2`,
			wantErr: false,
		},
		{
			name: "empty string included",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"", `value2`},
				lineNumber:    7,
				hasLineNumber: true,
				isFirst:       false,
			},
			want:    `7	-	value2`,
			wantErr: false,
		},
		{
			name: "with header 1",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "123"},
				lineNumber:    1,
				hasLineNumber: true,
				isFirst:       true,
			},
			want: `no	label1	label2
1	value1	123`,
			wantErr: false,
		},
		{
			name: "with header 2",
			args: args{
				labels:        []string{"label1", "label2"},
				values:        []string{"value1", "123"},
				lineNumber:    1,
				hasLineNumber: false,
				isFirst:       true,
			},
			want: `label1	label2
value1	123`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TSVLineHandler(tt.args.labels, tt.args.values, tt.args.lineNumber, tt.args.hasLineNumber, tt.args.isFirst)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func Test_addLineNumber(t *testing.T) {
	type args struct {
		labels     []string
		values     []string
		lineNumber int
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 []string
	}{
		{
			name: "basic",
			args: args{
				labels:     []string{"label1", "label2"},
				values:     []string{"value1", "value2"},
				lineNumber: 1,
			},
			want:  []string{"no", "label1", "label2"},
			want1: []string{"1", "value1", "value2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := addLineNumber(tt.args.labels, tt.args.values, tt.args.lineNumber)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got1, tt.want1)
			}
		})
	}
}
