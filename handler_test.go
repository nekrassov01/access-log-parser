package parser

import (
	"testing"
)

func TestJSONLineHandler(t *testing.T) {
	type args struct {
		labels  []string
		values  []string
		isFirst bool
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
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2"},
			},
			want:    `{"label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"\"value1", "value2"},
			},
			want:    `{"label1":"\"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2", "value3"},
			},
			want:    `{"label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1"},
			},
			want:    `{"label1":"value1"}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value 2"},
			},
			want:    `{"label1":"value1","label2":"value 2"}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"-", "\"-\""},
			},
			want:    `{"label1":"-","label2":"\"-\""}`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", `value\2`},
			},
			want:    `{"label1":"value1","label2":"value\\2"}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONLineHandler(tt.args.labels, tt.args.values, tt.args.isFirst)
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
		labels  []string
		values  []string
		isFirst bool
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
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2"},
			},
			want: `{
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"\"value1", "value2"},
			},
			want: `{
  "label1": "\"value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2", "value3"},
			},
			want: `{
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1"},
			},
			want: `{
  "label1": "value1"
}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value 2"},
			},
			want: `{
  "label1": "value1",
  "label2": "value 2"
}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"-", "\"-\""},
			},
			want: `{
  "label1": "-",
  "label2": "\"-\""
}`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", `value\2`},
			},
			want: `{
  "label1": "value1",
  "label2": "value\\2"
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSONLineHandler(tt.args.labels, tt.args.values, tt.args.isFirst)
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
		labels  []string
		values  []string
		isFirst bool
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
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2"},
			},
			want:    `label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"\"value1", "value2"},
			},
			want:    `label1="\"value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2", "value3"},
			},
			want:    `label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1"},
			},
			want:    `label1="value1"`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value 2"},
			},
			want:    `label1="value1" label2="value 2"`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"-", "\"-\""},
			},
			want:    `label1="-" label2="\"-\""`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", `value\2`},
			},
			want:    `label1="value1" label2="value\\2"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyValuePairLineHandler(tt.args.labels, tt.args.values, tt.args.isFirst)
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
		labels  []string
		values  []string
		isFirst bool
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
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2"},
			},
			want:    `label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"\"value1", "value2"},
			},
			want:    `label1:"value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value2", "value3"},
			},
			want:    `label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1"},
			},
			want:    `label1:value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", "value 2"},
			},
			want:    `label1:value1	label2:value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"-", "\"-\""},
			},
			want:    `label1:-	label2:"-"`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"value1", `value\2`},
			},
			want:    `label1:value1	label2:value\2`,
			wantErr: false,
		},
		{
			name: "empty string included",
			args: args{
				labels: []string{"label1", "label2"},
				values: []string{"", `value2`},
			},
			want:    `label1:-	label2:value2`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LTSVLineHandler(tt.args.labels, tt.args.values, tt.args.isFirst)
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
		labels  []string
		values  []string
		isFirst bool
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
				labels:  []string{"label1", "label2"},
				values:  []string{"value1", "123"},
				isFirst: false,
			},
			want:    `value1	123`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"\"value1", "123"},
				isFirst: false,
			},
			want:    `"value1	123`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"value1", "value2", "value3"},
				isFirst: false,
			},
			want:    `value1	value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"value1"},
				isFirst: false,
			},
			want:    `value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"value1", "value 2"},
				isFirst: false,
			},
			want:    `value1	value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"-", "\"-\""},
				isFirst: false,
			},
			want:    `-	"-"`,
			wantErr: false,
		},
		{
			name: "backslash included",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"value1", `value\2`},
				isFirst: false,
			},
			want:    `value1	value\2`,
			wantErr: false,
		},
		{
			name: "empty string included",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"", `value2`},
				isFirst: false,
			},
			want:    `-	value2`,
			wantErr: false,
		},
		{
			name: "with header",
			args: args{
				labels:  []string{"label1", "label2"},
				values:  []string{"value1", "123"},
				isFirst: true,
			},
			want: `label1	label2
value1	123`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TSVLineHandler(tt.args.labels, tt.args.values, tt.args.isFirst)
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
