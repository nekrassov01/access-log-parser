package parser

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nekrassov01/mintab"
)

// Result encapsulates the outcomes of parsing operations, detailing matched, unmatched, excluded,
// and skipped line counts, along with processing time and source information.
type Result struct {
	Total       int           `json:"total"`                // Total number of processed lines.
	Matched     int           `json:"matched"`              // Count of lines that matched the patterns.
	Unmatched   int           `json:"unmatched"`            // Count of lines that did not match any patterns.
	Excluded    int           `json:"excluded"`             // Count of lines excluded based on keyword search.
	Skipped     int           `json:"skipped"`              // Count of lines skipped explicitly.
	ElapsedTime time.Duration `json:"elapsedTime"`          // Processing time for the log data.
	Source      string        `json:"source"`               // Source of the log data.
	ZipEntries  []string      `json:"zipEntries,omitempty"` // List of processed zip entries, if applicable.
	Errors      []Errors      `json:"errors"`               // Collection of errors encountered during parsing.
	inputType   inputType     `json:"-"`                    // Type of input being processed.
}

// Errors stores information about log lines that couldn't be parsed
// according to the provided patterns. This helps in tracking and analyzing
// log lines that do not conform to expected formats.
type Errors struct {
	Entry      string `json:"entry,omitempty"` // Optional entry name if the log came from a zip file.
	LineNumber int    `json:"lineNumber"`      // Line number of the problematic log entry.
	Line       string `json:"line"`            // Content of the problematic log line.
}

// String generates a summary report of the parsing process,
// including a table of unmatched lines and a summary of counts.
func (r *Result) String() string {
	b := &strings.Builder{}
	top := 10
	cr := r.copy()
	omit := cr.beforeErrorsTable(top)
	var errTable, sumTable *mintab.Table
	var err error
	if len(cr.Errors) > 0 {
		errTable, err = cr.newErrorsTable(b)
		if err != nil {
			return err.Error()
		}
	}
	sumTable, err = cr.newSummaryTable(b)
	if err != nil {
		return err.Error()
	}
	if r.inputType == inputTypeStream {
		b.WriteString("\n")
	}
	b.WriteString("\n" + "\033[1;36m" + "/* SUMMARY */" + "\033[0m" + "\n\n")
	sumTable.Out()
	b.WriteString("\n" + "\033[2;37m")
	b.WriteString("Total     : Total number of log line processed\n")
	b.WriteString("Matched   : Number of log line that successfully matched pattern\n")
	b.WriteString("Unmatched : Number of log line that did not match any pattern\n")
	b.WriteString("Excluded  : Number of log line that did not hit by keyword search\n")
	b.WriteString("Skipped   : Number of log line that skipped by line number (disabled in stream mode)\n")
	b.WriteString("\033[0m")
	if errTable == nil {
		return b.String()
	}
	b.WriteString("\n" + "\033[1;36m" + "/* UNMATCH LINES */" + "\033[0m" + "\n\n")
	errTable.Out()
	if omit {
		b.WriteString("\033[33m" + fmt.Sprintf("// Show only the first %d of %d errors\n", top, len(r.Errors)) + "\033[0m")
	}
	b.WriteString("\n" + "\033[2;37m")
	b.WriteString("LineNumber : Line number of the log that did not match any pattern\n")
	b.WriteString("Line       : Raw log line that did not match any pattern\n")
	b.WriteString("\033[0m")
	return b.String()
}

// newSummaryTable creates a mintab.Table for summarizing the parsing results.
// It configures the table based on the type of input being processed.
func (r *Result) newSummaryTable(w io.Writer) (*mintab.Table, error) {
	var i []int
	switch r.inputType {
	case inputTypeStream:
		i = []int{4, 6, 7, 8, 9, 10}
	case inputTypeString:
		i = []int{6, 7, 8, 9, 10}
	case inputTypeFile, inputTypeGzip:
		i = []int{7, 8, 9, 10}
	case inputTypeZip:
		i = []int{8, 9, 10}
	default:
	}
	table := mintab.New(w, mintab.WithFormat(mintab.FormatText), mintab.WithIgnoreFields(i))
	r.Errors = []Errors{}
	if err := table.Load(r); err != nil {
		return nil, fmt.Errorf("cannot build result")
	}
	return table, nil
}

// newErrorsTable creates a mintab.Table specifically for displaying unmatched log lines.
// It adjusts the columns to ignore based on the input type.
func (r *Result) newErrorsTable(w io.Writer) (*mintab.Table, error) {
	var i []int
	switch r.inputType {
	case inputTypeZip:
		i = []int{}
	default:
		i = []int{0}
	}
	table := mintab.New(w, mintab.WithFormat(mintab.FormatText), mintab.WithIgnoreFields(i))
	if err := table.Load(r.Errors); err != nil {
		return nil, err
	}
	return table, nil
}

// beforeErrorsTable prepares the Errors slice for display, truncating if necessary
// and applying formatting to each line. It returns whether truncation occurred.
func (r *Result) beforeErrorsTable(n int) bool {
	elen := len(r.Errors)
	omit := elen > n
	if omit {
		r.Errors = r.Errors[:n]
	}
	for i, er := range r.Errors {
		er.Entry = fold(er.Entry, 18)
		er.Line = strings.ReplaceAll(fold(er.Line, 94), "\t", "\\t")
		r.Errors[i] = er
	}
	return omit
}

// copy creates a deep copy of the Result struct to avoid modifying
// the original data during the string formatting process.
func (r *Result) copy() Result {
	cr := *r
	cr.Errors = make([]Errors, len(r.Errors))
	copy(cr.Errors, r.Errors)
	return cr
}

// fold inserts line breaks into a string at specified intervals,
// used for formatting long lines.
func fold(s string, w int) string {
	b := &strings.Builder{}
	runes := []rune(s)
	for i, r := range runes {
		b.WriteRune(r)
		if (i+1)%w == 0 && i+1 < len(runes) {
			b.WriteRune('\n')
		}
	}
	return b.String()
}
