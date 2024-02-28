package parser

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nekrassov01/mintab"
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
				"\033[1;36m" +
				"/* SUMMARY */" +
				"\033[0m" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"|     1 |       1 |         0 |        0 |       0 | 1h0m0s      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+\n" +
				"\n" +
				"\033[2;37m" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not hit by keyword search\n" +
				"Skipped   : Number of log line that skipped by line number (disabled in stream mode)\n" +
				"\033[0m",
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
				"\033[1;36m" +
				"/* SUMMARY */" +
				"\033[0m" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime | Source   |\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"|     1 |       1 |         0 |        0 |       0 | 1h0m0s      | test.txt |\n" +
				"+-------+---------+-----------+----------+---------+-------------+----------+\n" +
				"\n" +
				"\033[2;37m" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not hit by keyword search\n" +
				"Skipped   : Number of log line that skipped by line number (disabled in stream mode)\n" +
				"\033[0m",
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
						Entry:      "2.log",
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
				"\033[1;36m" +
				"/* SUMMARY */" +
				"\033[0m" +
				"\n\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"| Total | Matched | Unmatched | Excluded | Skipped | ElapsedTime | Source  | ZipEntries |\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"|    13 |       1 |        12 |        0 |       0 | 1h0m0s      | 123.zip | 1.log      |\n" +
				"|       |         |           |          |         |             |         | 2.log      |\n" +
				"|       |         |           |          |         |             |         | 3.log      |\n" +
				"+-------+---------+-----------+----------+---------+-------------+---------+------------+\n" +
				"\n" +
				"\033[2;37m" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not hit by keyword search\n" +
				"Skipped   : Number of log line that skipped by line number (disabled in stream mode)\n" +
				"\033[0m" +
				"\n" +
				"\033[1;36m" +
				"/* UNMATCH LINES */" +
				"\033[0m" +
				"\n\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| Entry | LineNumber | Line                                                                                                 |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          2 | aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |\n" +
				"|       |            | aaaaaaaaaaaaaaaaaaaa                                                                                 |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          3 | aaa                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          4 | aaa                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          5 | aaa                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          6 | aaa                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 2.log |          7 | aaa                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 3.log |          2 | bbb                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 3.log |          3 | bbb                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 3.log |          4 | bbb                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"| 3.log |          5 | bbb                                                                                                  |\n" +
				"+-------+------------+------------------------------------------------------------------------------------------------------+\n" +
				"\033[33m" +
				"// Show only the first 10 of 12 errors\n" +
				"\033[0m" +
				"\n" +
				"\033[2;37m" +
				"LineNumber : Line number of the log that did not match any pattern\n" +
				"Line       : Raw log line that did not match any pattern\n" +
				"\033[0m",
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
				"\033[1;36m" +
				"/* SUMMARY */" +
				"\033[0m" +
				"\n\n" +
				"+-------+---------+-----------+----------+-------------+\n" +
				"| Total | Matched | Unmatched | Excluded | ElapsedTime |\n" +
				"+-------+---------+-----------+----------+-------------+\n" +
				"|     2 |       1 |         1 |        0 | 1h0m0s      |\n" +
				"+-------+---------+-----------+----------+-------------+\n" +
				"\n" +
				"\033[2;37m" +
				"Total     : Total number of log line processed\n" +
				"Matched   : Number of log line that successfully matched pattern\n" +
				"Unmatched : Number of log line that did not match any pattern\n" +
				"Excluded  : Number of log line that did not hit by keyword search\n" +
				"Skipped   : Number of log line that skipped by line number (disabled in stream mode)\n" +
				"\033[0m" +
				"\n" +
				"\033[1;36m" +
				"/* UNMATCH LINES */" +
				"\033[0m" +
				"\n\n" +
				"+------------+------+\n" +
				"| LineNumber | Line |\n" +
				"+------------+------+\n" +
				"|          2 | aaa  |\n" +
				"+------------+------+\n" +
				"\n" +
				"\033[2;37m" +
				"LineNumber : Line number of the log that did not match any pattern\n" +
				"Line       : Raw log line that did not match any pattern\n" +
				"\033[0m",
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

func TestResult_newSummaryTable(t *testing.T) {
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
		name    string
		fields  fields
		want    *mintab.Table
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
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
			w := &bytes.Buffer{}
			got, err := r.newSummaryTable(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Result.newSummaryTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Result.newSummaryTable() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Result.newSummaryTable() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestResult_newErrorsTable(t *testing.T) {
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
		name    string
		fields  fields
		want    *mintab.Table
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
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
			w := &bytes.Buffer{}
			got, err := r.newErrorsTable(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Result.newErrorsTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Result.newErrorsTable() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Result.newErrorsTable() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestResult_beforeErrorsTable(t *testing.T) {
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
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
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
			if got := r.beforeErrorsTable(tt.args.n); got != tt.want {
				t.Errorf("Result.beforeErrorsTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_copy(t *testing.T) {
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
		want   Result
	}{
		// TODO: Add test cases.
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
			if got := r.copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Result.copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fold(t *testing.T) {
	type args struct {
		s string
		w int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fold(tt.args.s, tt.args.w); got != tt.want {
				t.Errorf("fold() = %v, want %v", got, tt.want)
			}
		})
	}
}
