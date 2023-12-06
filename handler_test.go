package parser

import (
	"fmt"
	"testing"
)

func TestJSONLineHandler(t *testing.T) {
	type args struct {
		values   []string
		labels   []string
		index    int
		hasIndex bool
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
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2"},
				index:    1,
				hasIndex: true,
			},
			want:    `{"index":"1","label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "val\"ue2"},
				index:    2,
				hasIndex: true,
			},
			want:    `{"index":"2","label1":"value1","label2":"val\"ue2"}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2", "value3"},
				index:    3,
				hasIndex: true,
			},
			want:    `{"index":"3","label1":"value1","label2":"value2"}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1"},
				index:    4,
				hasIndex: true,
			},
			want:    `{"index":"4","label1":"value1"}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value 2"},
				index:    5,
				hasIndex: true,
			},
			want:    `{"index":"5","label1":"value1","label2":"value 2"}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"-", "\"-\""},
				index:    6,
				hasIndex: true,
			},
			want:    `{"index":"6","label1":"-","label2":"\"-\""}`,
			wantErr: false,
		},
		{
			name: "slash included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", `value\2`},
				index:    7,
				hasIndex: true,
			},
			want:    `{"index":"7","label1":"value1","label2":"value\\2"}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONLineHandler(tt.args.labels, tt.args.values, tt.args.index, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONLineHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSONLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				m: regexAllMatchMetadata,
			},
			want:    fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				m: regexContainsUnmatchMetadata,
			},
			want:    fmt.Sprintf(regexContainsUnmatchMetadataSerialized, ""),
			wantErr: false,
		},
		{
			name: "contains skip",
			args: args{
				m: regexContainsSkipMetadata,
			},
			want:    fmt.Sprintf(regexContainsSkipMetadataSerialized, ""),
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				m: regexAllUnmatchMetadata,
			},
			want:    fmt.Sprintf(regexAllUnmatchMetadataSerialized, ""),
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				m: regexAllSkipMetadata,
			},
			want:    fmt.Sprintf(regexAllSkipMetadataSerialized, ""),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONMetadataHandler(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONMetadataHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSONMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrettyJSONLineHandler(t *testing.T) {
	type args struct {
		values   []string
		labels   []string
		index    int
		hasIndex bool
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
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2"},
				index:    1,
				hasIndex: true,
			},
			want: `{
  "index": "1",
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "val\"ue2"},
				index:    2,
				hasIndex: true,
			},
			want: `{
  "index": "2",
  "label1": "value1",
  "label2": "val\"ue2"
}`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2", "value3"},
				index:    3,
				hasIndex: true,
			},
			want: `{
  "index": "3",
  "label1": "value1",
  "label2": "value2"
}`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1"},
				index:    4,
				hasIndex: true,
			},
			want: `{
  "index": "4",
  "label1": "value1"
}`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value 2"},
				index:    5,
				hasIndex: true,
			},
			want: `{
  "index": "5",
  "label1": "value1",
  "label2": "value 2"
}`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"-", "\"-\""},
				index:    6,
				hasIndex: true,
			},
			want: `{
  "index": "6",
  "label1": "-",
  "label2": "\"-\""
}`,
			wantErr: false,
		},
		{
			name: "slash included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", `value\2`},
				index:    7,
				hasIndex: true,
			},
			want: `{
  "index": "7",
  "label1": "value1",
  "label2": "value\\2"
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSONLineHandler(tt.args.labels, tt.args.values, tt.args.index, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSONLineHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrettyJSONLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrettyJSONMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				m: regexAllMatchMetadata,
			},
			want: `{
  "total": 5,
  "matched": 5,
  "unmatched": 0,
  "skipped": 0,
  "source": "",
  "errors": null
}`,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				m: regexContainsUnmatchMetadata,
			},
			want: `{
  "total": 5,
  "matched": 4,
  "unmatched": 1,
  "skipped": 0,
  "source": "",
  "errors": [
    {
      "index": 4,
      "record": "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""
    }
  ]
}`,
			wantErr: false,
		},
		{
			name: "contains skip",
			args: args{
				m: regexContainsSkipMetadata,
			},
			want: `{
  "total": 5,
  "matched": 3,
  "unmatched": 0,
  "skipped": 2,
  "source": "",
  "errors": null
}`,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				m: regexAllUnmatchMetadata,
			},
			want: `{
  "total": 5,
  "matched": 0,
  "unmatched": 5,
  "skipped": 0,
  "source": "",
  "errors": [
    {
      "index": 1,
      "record": "a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""
    },
    {
      "index": 2,
      "record": "3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""
    },
    {
      "index": 3,
      "record": "8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"
    },
    {
      "index": 4,
      "record": "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"
    },
    {
      "index": 5,
      "record": "01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"
    }
  ]
}`,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				m: regexAllSkipMetadata,
			},
			want: `{
  "total": 5,
  "matched": 0,
  "unmatched": 0,
  "skipped": 5,
  "source": "",
  "errors": null
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyJSONMetadataHandler(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSONMetadataHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrettyJSONMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyValuePairLineHandler(t *testing.T) {
	type args struct {
		values   []string
		labels   []string
		index    int
		hasIndex bool
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
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2"},
				index:    1,
				hasIndex: true,
			},
			want:    `index="1" label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "val\"ue2"},
				index:    2,
				hasIndex: true,
			},
			want:    `index="2" label1="value1" label2="val\"ue2"`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2", "value3"},
				index:    3,
				hasIndex: true,
			},
			want:    `index="3" label1="value1" label2="value2"`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1"},
				index:    4,
				hasIndex: true,
			},
			want:    `index="4" label1="value1"`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value 2"},
				index:    5,
				hasIndex: true,
			},
			want:    `index="5" label1="value1" label2="value 2"`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"-", "\"-\""},
				index:    6,
				hasIndex: true,
			},
			want:    `index="6" label1="-" label2="\"-\""`,
			wantErr: false,
		},
		{
			name: "slash included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", `value\2`},
				index:    7,
				hasIndex: true,
			},
			want:    `index="7" label1="value1" label2="value\\2"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyValuePairLineHandler(tt.args.labels, tt.args.values, tt.args.index, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyValuePairLineHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KeyValuePairLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyValuePairMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				m: regexAllMatchMetadata,
			},
			want:    `total=5 matched=5 unmatched=0 skipped=0 source="" errors=null`,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				m: regexContainsUnmatchMetadata,
			},
			want:    `total=5 matched=4 unmatched=1 skipped=0 source="" errors=[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]`,
			wantErr: false,
		},
		{
			name: "contains skip",
			args: args{
				m: regexContainsSkipMetadata,
			},
			want:    `total=5 matched=3 unmatched=0 skipped=2 source="" errors=null`,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				m: regexAllUnmatchMetadata,
			},
			want:    `total=5 matched=0 unmatched=5 skipped=0 source="" errors=[{"index":1,"record":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""},{"index":2,"record":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""},{"index":3,"record":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"},{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"},{"index":5,"record":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"}]`,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				m: regexAllSkipMetadata,
			},
			want:    `total=5 matched=0 unmatched=0 skipped=5 source="" errors=null`,
			wantErr: false,
		},
		{
			name: "source is not null",
			args: args{
				m: &Metadata{
					Total:     5,
					Matched:   5,
					Unmatched: 0,
					Skipped:   0,
					Source:    "test.log",
					Errors:    nil,
				},
			},
			want:    `total=5 matched=5 unmatched=0 skipped=0 source="test.log" errors=null`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyValuePairMetadataHandler(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyValuePairMetadataHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KeyValuePairMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVLineHandler(t *testing.T) {
	type args struct {
		values   []string
		labels   []string
		index    int
		hasIndex bool
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
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2"},
				index:    1,
				hasIndex: true,
			},
			want:    `index:1	label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "val\"ue2"},
				index:    2,
				hasIndex: true,
			},
			want:    `index:2	label1:value1	label2:val"ue2`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2", "value3"},
				index:    3,
				hasIndex: true,
			},
			want:    `index:3	label1:value1	label2:value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1"},
				index:    4,
				hasIndex: true,
			},
			want:    `index:4	label1:value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value 2"},
				index:    5,
				hasIndex: true,
			},
			want:    `index:5	label1:value1	label2:value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"-", "\"-\""},
				index:    6,
				hasIndex: true,
			},
			want:    `index:6	label1:-	label2:"-"`,
			wantErr: false,
		},
		{
			name: "slash included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", `value\2`},
				index:    7,
				hasIndex: true,
			},
			want:    `index:7	label1:value1	label2:value\2`,
			wantErr: false,
		},
		{
			name: "blank included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", ""},
				index:    8,
				hasIndex: true,
			},
			want:    `index:8	label1:value1	label2:-`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LTSVLineHandler(tt.args.labels, tt.args.values, tt.args.index, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVLineHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LTSVLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLTSVMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				m: regexAllMatchMetadata,
			},
			want:    `total:5	matched:5	unmatched:0	skipped:0	source:-	errors:null`,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				m: regexContainsUnmatchMetadata,
			},
			want:    `total:5	matched:4	unmatched:1	skipped:0	source:-	errors:[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]`,
			wantErr: false,
		},
		{
			name: "contains skip",
			args: args{
				m: regexContainsSkipMetadata,
			},
			want:    `total:5	matched:3	unmatched:0	skipped:2	source:-	errors:null`,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				m: regexAllUnmatchMetadata,
			},
			want:    `total:5	matched:0	unmatched:5	skipped:0	source:-	errors:[{"index":1,"record":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""},{"index":2,"record":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""},{"index":3,"record":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"},{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"},{"index":5,"record":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"}]`,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				m: regexAllSkipMetadata,
			},
			want:    `total:5	matched:0	unmatched:0	skipped:5	source:-	errors:null`,
			wantErr: false,
		},
		{
			name: "source is not null",
			args: args{
				m: &Metadata{
					Total:     5,
					Matched:   5,
					Unmatched: 0,
					Skipped:   0,
					Source:    "test.log",
					Errors:    nil,
				},
			},
			want:    `total:5	matched:5	unmatched:0	skipped:0	source:test.log	errors:null`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LTSVMetadataHandler(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("LTSVMetadataHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LTSVMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTSVLineHandler(t *testing.T) {
	type args struct {
		values   []string
		labels   []string
		index    int
		hasIndex bool
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
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2"},
				index:    1,
				hasIndex: true,
			},
			want:    `1	value1	value2`,
			wantErr: false,
		},
		{
			name: "invalid json character",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "val\"ue2"},
				index:    2,
				hasIndex: true,
			},
			want:    `2	value1	val"ue2`,
			wantErr: false,
		},
		{
			name: "more matches than fields",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value2", "value3"},
				index:    3,
				hasIndex: true,
			},
			want:    `3	value1	value2`,
			wantErr: false,
		},
		{
			name: "more fields than matches",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1"},
				index:    4,
				hasIndex: true,
			},
			want:    `4	value1`,
			wantErr: false,
		},
		{
			name: "space included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", "value 2"},
				index:    5,
				hasIndex: true,
			},
			want:    `5	value1	value 2`,
			wantErr: false,
		},
		{
			name: "hyphen included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"-", "\"-\""},
				index:    6,
				hasIndex: true,
			},
			want:    `6	-	"-"`,
			wantErr: false,
		},
		{
			name: "slash included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", `value\2`},
				index:    7,
				hasIndex: true,
			},
			want:    `7	value1	value\2`,
			wantErr: false,
		},
		{
			name: "blank included",
			args: args{
				labels:   []string{"label1", "label2"},
				values:   []string{"value1", ""},
				index:    8,
				hasIndex: true,
			},
			want:    `8	value1	-`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TSVLineHandler(tt.args.labels, tt.args.values, tt.args.index, tt.args.hasIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("TSVLineHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TSVLineHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTSVMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				m: regexAllMatchMetadata,
			},
			want:    `5	5	0	0	-	null`,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				m: regexContainsUnmatchMetadata,
			},
			want:    `5	4	1	0	-	[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]`,
			wantErr: false,
		},
		{
			name: "contains skip",
			args: args{
				m: regexContainsSkipMetadata,
			},
			want:    `5	3	0	2	-	null`,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				m: regexAllUnmatchMetadata,
			},
			want:    `5	0	5	0	-	[{"index":1,"record":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""},{"index":2,"record":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""},{"index":3,"record":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"},{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"},{"index":5,"record":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"}]`,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				m: regexAllSkipMetadata,
			},
			want:    `5	0	0	5	-	null`,
			wantErr: false,
		},
		{
			name: "source is not null",
			args: args{
				m: &Metadata{
					Total:     5,
					Matched:   5,
					Unmatched: 0,
					Skipped:   0,
					Source:    "test.log",
					Errors:    nil,
				},
			},
			want:    `5	5	0	0	test.log	null`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TSVMetadataHandler(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("TSVMetadataHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TSVMetadataHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
