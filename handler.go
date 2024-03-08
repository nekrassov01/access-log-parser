package parser

import (
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
)

const lineNumberLabel = "no"

// JSONLineHandler serializes log lines into JSON (NDJSON) format. It keywords the line number if specified.
// Labels and values are combined into key-value pairs, and the result is a single JSON object.
func JSONLineHandler(labels []string, values []string, lineNumber int, hasLineNumber, _ bool) (string, error) {
	if hasLineNumber {
		labels, values = addLineNumber(labels, values, lineNumber)
	}
	b := &strings.Builder{}
	b.WriteString("{")
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\"")
			b.WriteString(labels[i])
			b.WriteString("\":")
			b.WriteString(strconv.Quote(value))
		}
	}
	b.WriteString("}")
	return b.String(), nil
}

// PrettyJSONLineHandler enhances JSONLineHandler by formatting the output for readability. It uses indentation and new lines.
func PrettyJSONLineHandler(labels []string, values []string, lineNumber int, hasLineNumber, _ bool) (string, error) {
	if hasLineNumber {
		labels, values = addLineNumber(labels, values, lineNumber)
	}
	b := &strings.Builder{}
	b.WriteString("{\n")
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString("  \"")
			b.WriteString(labels[i])
			b.WriteString("\": ")
			b.WriteString(strconv.Quote(value))
		}
	}
	b.WriteString("\n}")
	return b.String(), nil
}

// KeyValuePairLineHandler converts log lines into a space-separated string of key-value pairs.
func KeyValuePairLineHandler(labels []string, values []string, lineNumber int, hasLineNumber, _ bool) (string, error) {
	if hasLineNumber {
		labels, values = addLineNumber(labels, values, lineNumber)
	}
	b := &strings.Builder{}
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(labels[i])
			b.WriteString("=")
			b.WriteString(strconv.Quote(value))
		}
	}
	return b.String(), nil
}

// LTSVLineHandler formats log lines as LTSV (Labeled Tab-separated Values).
func LTSVLineHandler(labels []string, values []string, lineNumber int, hasLineNumber, _ bool) (string, error) {
	if hasLineNumber {
		labels, values = addLineNumber(labels, values, lineNumber)
	}
	b := &strings.Builder{}
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				b.WriteString("\t")
			}
			b.WriteString(labels[i])
			b.WriteString(":")
			if value == "" {
				b.WriteString("-")
			} else {
				b.WriteString(value)
			}
		}
	}
	return b.String(), nil
}

// TSVLineHandler formats log lines as TSV (Tab-separated Values).
func TSVLineHandler(labels []string, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
	if hasLineNumber {
		labels, values = addLineNumber(labels, values, lineNumber)
	}
	b := &strings.Builder{}
	if isFirst {
		header := strings.Join(labels, "\t")
		if isatty.IsTerminal(os.Stdout.Fd()) {
			header = "\033[1;37m" + header + "\033[0m"
		}
		b.WriteString(header)
		b.WriteString("\n")
	}
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				b.WriteString("\t")
			}
			if value == "" {
				b.WriteString("-")
			} else {
				b.WriteString(value)
			}
		}
	}
	return b.String(), nil
}

// addLineNumber prepends the line number to labels and values.
func addLineNumber(labels []string, values []string, lineNumber int) ([]string, []string) {
	return append([]string{lineNumberLabel}, labels...), append([]string{strconv.Itoa(lineNumber)}, values...)
}
