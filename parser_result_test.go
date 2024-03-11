package parser

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestResult_String(t *testing.T) {
	type fields struct {
		Total       int
		Matched     int
		Unmatched   int
		Excluded    int
		Skipped     int
		ElapsedTime time.Duration
		Source      string
		ZipEntries  []string
		Errors      []Errors
		inputType   inputType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "basic",
			fields: fields{
				Total:       1,
				Matched:     1,
				Unmatched:   0,
				Excluded:    0,
				Skipped:     0,
				ElapsedTime: time.Hour,
				Source:      "",
				ZipEntries:  nil,
				Errors:      []Errors{},
				inputType:   inputTypeString,
			},
			want: "\n" +
				"/* SUMMARY */" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"|     1 |       1 |         0 |        0 |       0 | 1h0m0s      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"\n" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not extract by filter expressions\n" +
				"Skipped   : Number of log line that skipped by line number\n",
		},
		{
			name: "file",
			fields: fields{
				Total:       1,
				Matched:     1,
				Unmatched:   0,
				Excluded:    0,
				Skipped:     0,
				ElapsedTime: time.Hour,
				Source:      "test.txt",
				ZipEntries:  nil,
				Errors:      []Errors{},
				inputType:   inputTypeFile,
			},
			want: "\n" +
				"/* SUMMARY */" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime | Source   |\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"|     1 |       1 |         0 |        0 |       0 | 1h0m0s      | test.txt |\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"\n" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not extract by filter expressions\n" +
				"Skipped   : Number of log line that skipped by line number\n",
		},
		{
			name: "all",
			fields: fields{
				Total:       13,
				Matched:     1,
				Unmatched:   12,
				Excluded:    0,
				Skipped:     0,
				ElapsedTime: time.Hour,
				Source:      "123.zip",
				ZipEntries:  []string{"1.log", "2.log", "3.log"},
				Errors: []Errors{
					{
						Entry:      "2.log",
						LineNumber: 2,
						Line:       strings.Repeat("a", 120),
					},
					{
						Entry:      strings.Repeat("a", 20),
						LineNumber: 3,
						Line:       "aaa",
					},
					{
						Entry:      "2.log",
						LineNumber: 4,
						Line:       "aaa",
					},
					{
						Entry:      "2.log",
						LineNumber: 5,
						Line:       "aaa",
					},
					{
						Entry:      "2.log",
						LineNumber: 6,
						Line:       "aaa",
					},
					{
						Entry:      "2.log",
						LineNumber: 7,
						Line:       "aaa",
					},
					{
						Entry:      "3.log",
						LineNumber: 2,
						Line:       "bbb",
					},
					{
						Entry:      "3.log",
						LineNumber: 3,
						Line:       "bbb",
					},
					{
						Entry:      "3.log",
						LineNumber: 4,
						Line:       "bbb",
					},
					{
						Entry:      "3.log",
						LineNumber: 5,
						Line:       "bbb",
					},
					{
						Entry:      "3.log",
						LineNumber: 6,
						Line:       "bbb",
					},
					{
						Entry:      "3.log",
						LineNumber: 7,
						Line:       "bbb",
					},
				},
				inputType: inputTypeZip,
			},
			want: "\n" +
				"/* SUMMARY */" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime | Source  | ZipEntries |\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"|    13 |       1 |        12 |        0 |       0 | 1h0m0s      | 123.zip | 1.log      |\n" +
				"|       |         |           |          |         |             |         | 2.log      |\n" +
				"|       |         |           |          |         |             |         | 3.log      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"\n" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not extract by filter expressions\n" +
				"Skipped   : Number of log line that skipped by line number\n" +
				"\n" +
				"/* UNMATCH LINES */" +
				"\n\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| Entry              | LineNumber | Line                                                                                           |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 2.log              |          2 | aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |\n" +
				"|                    |            | aaaaaaaaaaaaaaaaaaaaaaaaaa                                                                     |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| aaaaaaaaaaaaaaaaaa |          3 | aaa                                                                                            |\n" +
				"| aa                 |            |                                                                                                |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 2.log              |          4 | aaa                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 2.log              |          5 | aaa                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 2.log              |          6 | aaa                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 2.log              |          7 | aaa                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 3.log              |          2 | bbb                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 3.log              |          3 | bbb                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 3.log              |          4 | bbb                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"| 3.log              |          5 | bbb                                                                                            |\n" +
				"+--------------------+------------+------------------------------------------------------------------------------------------------+\n" +
				"// Show only the first 10 of 12 errors\n" +
				"\n" +
				"LineNumber : Line number of the log that did not match any pattern\n" +
				"Line       : Raw log line that did not match any pattern\n",
		},
		{
			name: "stream",
			fields: fields{
				Total:       2,
				Matched:     1,
				Unmatched:   1,
				Excluded:    0,
				Skipped:     0,
				ElapsedTime: time.Hour,
				Source:      "",
				ZipEntries:  nil,
				Errors: []Errors{
					{
						LineNumber: 2,
						Line:       "aaa",
					},
				},
				inputType: inputTypeStream,
			},
			want: "\n\n" +
				"/* SUMMARY */" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"|     2 |       1 |         1 |        0 |       0 | 1h0m0s      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"\n" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not extract by filter expressions\n" +
				"Skipped   : Number of log line that skipped by line number\n" +
				"\n" +
				"/* UNMATCH LINES */" +
				"\n\n" +
				"+------+\n" +
				"| Line |\n" +
				"+------+\n" +
				"| aaa  |\n" +
				"+------+\n" +
				"\n" +
				"LineNumber : Line number of the log that did not match any pattern\n" +
				"Line       : Raw log line that did not match any pattern\n",
		},
		{
			name: "string",
			fields: fields{
				Total:       2,
				Matched:     1,
				Unmatched:   1,
				Excluded:    0,
				Skipped:     0,
				ElapsedTime: time.Hour,
				Source:      "",
				ZipEntries:  nil,
				Errors: []Errors{
					{
						LineNumber: 2,
						Line:       "aaa",
					},
				},
				inputType: inputTypeString,
			},
			want: "\n" +
				"/* SUMMARY */" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"|     2 |       1 |         1 |        0 |       0 | 1h0m0s      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"\n" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not extract by filter expressions\n" +
				"Skipped   : Number of log line that skipped by line number\n" +
				"\n" +
				"/* UNMATCH LINES */" +
				"\n\n" +
				"+------------+------+\n" +
				"| LineNumber | Line |\n" +
				"+------------+------+\n" +
				"|          2 | aaa  |\n" +
				"+------------+------+\n" +
				"\n" +
				"LineNumber : Line number of the log that did not match any pattern\n" +
				"Line       : Raw log line that did not match any pattern\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				Total:       tt.fields.Total,
				Matched:     tt.fields.Matched,
				Unmatched:   tt.fields.Unmatched,
				Excluded:    tt.fields.Excluded,
				Skipped:     tt.fields.Skipped,
				ElapsedTime: tt.fields.ElapsedTime,
				Source:      tt.fields.Source,
				ZipEntries:  tt.fields.ZipEntries,
				Errors:      tt.fields.Errors,
				inputType:   tt.fields.inputType,
			}
			if diff := cmp.Diff(r.String(), tt.want); diff != "" {
				t.Errorf(diff)
			}
			if !reflect.DeepEqual(r.String(), tt.want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", r.String(), tt.want)
			}
		})
	}
}
