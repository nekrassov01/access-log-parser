package parser

import (
	"bytes"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

const size = 2024

// JSONLineHandler serializes log lines into JSON (NDJSON) format. It keywords the line number if specified.
// Labels and values are combined into key-value pairs, and the result is a single JSON object.
func JSONLineHandler(labels, values []string, _ bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.Grow(size)
	buf.WriteByte('{')
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('"')
			buf.WriteString(labels[i])
			buf.WriteString("\":")
			buf.WriteByte('"')
			writeEscapedString(buf, value)
			buf.WriteByte('"')
		}
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

// PrettyJSONLineHandler enhances JSONLineHandler by formatting the output for readability. It uses indentation and new lines.
func PrettyJSONLineHandler(labels, values []string, _ bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.Grow(size)
	buf.WriteString("{\n")
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				buf.WriteString(",\n")
			}
			buf.WriteString("  \"")
			buf.WriteString(labels[i])
			buf.WriteString("\": ")
			buf.WriteByte('"')
			writeEscapedString(buf, value)
			buf.WriteByte('"')
		}
	}
	buf.WriteString("\n}")
	return buf.String(), nil
}

// KeyValuePairLineHandler converts log lines into a space-separated string of key-value pairs.
func KeyValuePairLineHandler(labels, values []string, _ bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.Grow(size)
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(labels[i])
			buf.WriteByte('=')
			buf.WriteByte('"')
			writeEscapedString(buf, value)
			buf.WriteByte('"')
		}
	}
	return buf.String(), nil
}

// LTSVLineHandler formats log lines as LTSV (Labeled Tab-separated Values).
func LTSVLineHandler(labels, values []string, _ bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.Grow(size)
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				buf.WriteByte('\t')
			}
			buf.WriteString(labels[i])
			buf.WriteByte(':')
			if value == "" {
				buf.WriteByte('-')
			} else {
				buf.WriteString(value)
			}
		}
	}
	return buf.String(), nil
}

// TSVLineHandler formats log lines as TSV (Tab-separated Values).
func TSVLineHandler(labels, values []string, isFirst bool) (string, error) {
	buf := &bytes.Buffer{}
	buf.Grow(size)
	if isFirst {
		header := strings.Join(labels, "\t")
		if isatty.IsTerminal(os.Stdout.Fd()) {
			header = "\033[1;37m" + header + "\033[0m"
		}
		buf.WriteString(header)
		buf.WriteByte('\n')
	}
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				buf.WriteByte('\t')
			}
			if value == "" {
				buf.WriteByte('-')
			} else {
				buf.WriteString(value)
			}
		}
	}
	return buf.String(), nil
}

// EscapedString writes the string s to the given bytes.Buffer while properly escaping
// special characters (backslash, double quote, newline, carriage return, tab).
func writeEscapedString(buf *bytes.Buffer, s string) {
	for _, r := range s {
		switch r {
		case '\\':
			buf.WriteString("\\\\")
		case '"':
			buf.WriteString("\\\"")
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		default:
			buf.WriteRune(r)
		}
	}
}
