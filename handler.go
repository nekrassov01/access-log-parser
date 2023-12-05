package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSONLineHandler converts each log line to a JSON (NDJSON) format.
// It serializes the matched parts of the log line into a JSON object,
// with the log line index and matched values as key-value pairs.
func JSONLineHandler(labels []string, values []string, index int, hasIndex bool) (string, error) {
	if hasIndex {
		labels, values = addIndex(labels, values, index)
	}
	var builder strings.Builder
	builder.WriteString("{")
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				builder.WriteString(",")
			}
			builder.WriteString("\"")
			builder.WriteString(labels[i])
			builder.WriteString("\":")
			builder.WriteString(strconv.Quote(value))
		}
	}
	builder.WriteString("}")
	return builder.String(), nil
}

// JSONMetadataHandler is the default handler that converts  metadata to NDJSON format.
// It is used if no specific MetadataHandler is provided when the Parser is instantiated.
func JSONMetadataHandler(m *Metadata) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("cannot marshal result: %w", err)
	}
	return string(b), nil
}

// PrettyJSONLineHandler is similar to JSONLineHandler but formats the output in a human-readable JSON format.
// It first converts the line to a JSON string and then formats it with indents for better readability.
func PrettyJSONLineHandler(labels []string, values []string, index int, hasIndex bool) (string, error) {
	s, err := JSONLineHandler(labels, values, index, hasIndex)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}

// PrettyJSONMetadataHandler is similar to JSONMetadataHandler but formats the output in a human-readable JSON format.
// It first converts the metadata to a JSON string and then formats it with indents for better readability.
func PrettyJSONMetadataHandler(m *Metadata) (string, error) {
	s, err := JSONMetadataHandler(m)
	if err != nil {
		return "", err
	}
	return prettyJSON(s)
}

// prettyJSON takes a JSON string and formats it for better readability.
// It adds indents to the JSON string for a pretty print format.
func prettyJSON(s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return "", fmt.Errorf("cannot format string: %w", err)
	}
	return buf.String(), nil
}

// KeyValuePairLineHandler converts log lines into a key-value pair format.
// Each matched log line is represented as a string of key-value pairs, separated by spaces.
func KeyValuePairLineHandler(labels []string, values []string, index int, hasIndex bool) (string, error) {
	if hasIndex {
		labels, values = addIndex(labels, values, index)
	}
	var builder strings.Builder
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(labels[i])
			builder.WriteString("=")
			builder.WriteString(strconv.Quote(value))
		}
	}
	return builder.String(), nil
}

// KeyValuePairMetadataHandler converts the metadata into a string of key-value pairs.
// It serializes metadata like total line count, matched count, etc., into a single string.
func KeyValuePairMetadataHandler(m *Metadata) (string, error) {
	var builder strings.Builder
	builder.WriteString("total=")
	builder.WriteString(strconv.Itoa(m.Total))
	builder.WriteString(" matched=")
	builder.WriteString(strconv.Itoa(m.Matched))
	builder.WriteString(" unmatched=")
	builder.WriteString(strconv.Itoa(m.Unmatched))
	builder.WriteString(" skipped=")
	builder.WriteString(strconv.Itoa(m.Skipped))
	builder.WriteString(" source=")
	builder.WriteString(strconv.Quote(m.Source))
	builder.WriteString(" errors=")
	e, err := json.Marshal(m.Errors)
	if err != nil {
		return "", fmt.Errorf("cannot marshal errors: %w", err)
	}
	builder.Write(e)
	return builder.String(), nil
}

// LTSVLineHandler formats each log line as LTSV (Labeled Tab-separated Values).
// It converts each matched log line into a tab-separated string, with each field as a label-value pair.
func LTSVLineHandler(labels []string, values []string, index int, hasIndex bool) (string, error) {
	if hasIndex {
		labels, values = addIndex(labels, values, index)
	}
	var builder strings.Builder
	for i, value := range values {
		if i < len(labels) {
			if i > 0 {
				builder.WriteString("\t")
			}
			builder.WriteString(labels[i])
			builder.WriteString(":")
			if value == "" {
				builder.WriteString("-")
			} else {
				builder.WriteString(value)
			}
		}
	}
	return builder.String(), nil
}

// LTSVMetadataHandler formats the metadata as LTSV.
// It serializes metadata into a tab-separated string, similar to LTSV format, for each metadata field.
func LTSVMetadataHandler(m *Metadata) (string, error) {
	var builder strings.Builder
	builder.WriteString("total:")
	builder.WriteString(strconv.Itoa(m.Total))
	builder.WriteString("\tmatched:")
	builder.WriteString(strconv.Itoa(m.Matched))
	builder.WriteString("\tunmatched:")
	builder.WriteString(strconv.Itoa(m.Unmatched))
	builder.WriteString("\tskipped:")
	builder.WriteString(strconv.Itoa(m.Skipped))
	builder.WriteString("\tsource:")
	if m.Source == "" {
		builder.WriteString("-")
	} else {
		builder.WriteString(m.Source)
	}
	builder.WriteString("\terrors:")
	e, err := json.Marshal(m.Errors)
	if err != nil {
		return "", fmt.Errorf("cannot marshal errors: %w", err)
	}
	builder.Write(e)
	return builder.String(), nil
}

// addIndex adds a line number to the beginning of the label and value slices.
func addIndex(labels []string, values []string, index int) ([]string, []string) {
	return append([]string{"index"}, labels...), append([]string{strconv.Itoa(index)}, values...)
}
