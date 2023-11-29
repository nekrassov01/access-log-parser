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
func JSONLineHandler(values []string, labels []string, index int) (string, error) {
	var builder strings.Builder
	builder.WriteString("{\"index\":")
	builder.WriteString(strconv.Itoa(index))
	for i, value := range values {
		if i < len(labels) {
			builder.WriteString(",\"")
			builder.WriteString(labels[i])
			if value == "\"-\"" {
				builder.WriteString("\":\"-\"")
				continue
			}
			builder.WriteString("\":")
			b, err := json.Marshal(value)
			if err != nil {
				return "", fmt.Errorf("cannot marshal matched string \"%s\" as json: %w", value, err)
			}
			builder.Write(b)
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
		return "", fmt.Errorf("cannot marshal result as json: %w", err)
	}
	return string(b), nil
}

// PrettyJSONLineHandler is similar to JSONLineHandler but formats the output in a human-readable JSON format.
// It first converts the line to a JSON string and then formats it with indents for better readability.
func PrettyJSONLineHandler(values []string, labels []string, index int) (string, error) {
	s, err := JSONLineHandler(values, labels, index)
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
		return "", fmt.Errorf("cannot format string as json: %w", err)
	}
	return buf.String(), nil
}

// KeyValuePairLineHandler converts log lines into a key-value pair format.
// Each matched log line is represented as a string of key-value pairs, separated by spaces.
func KeyValuePairLineHandler(values []string, labels []string, index int) (string, error) {
	var builder strings.Builder
	builder.WriteString("index=")
	builder.WriteString(strconv.Itoa(index))
	for i, value := range values {
		if i < len(labels) {
			builder.WriteString(" ")
			builder.WriteString(labels[i])
			builder.WriteString(`="`)
			if value == "\"-\"" {
				builder.WriteString("-")
			} else {
				for _, v := range value {
					if v == '"' {
						builder.WriteString(`\"`)
					} else {
						builder.WriteRune(v)
					}
				}
			}
			builder.WriteString(`"`)
		}
	}
	return builder.String(), nil
}

// KeyValuePairMetadataHandler converts the metadata into a string of key-value pairs.
// It serializes metadata like total line count, matched count, etc., into a single string.
func KeyValuePairMetadataHandler(m *Metadata) (string, error) {
	e, err := json.Marshal(m.Errors)
	if err != nil {
		return "", fmt.Errorf("cannot marshal errors as json: %w", err)
	}
	return fmt.Sprintf(
		"total=%d matched=%d unmatched=%d skipped=%d source=\"%s\" errors=%s",
		m.Total, m.Matched, m.Unmatched, m.Skipped, m.Source, e,
	), nil
}

// LTSVLineHandler formats each log line as LTSV (Labeled Tab-separated Values).
// It converts each matched log line into a tab-separated string, with each field as a label-value pair.
func LTSVLineHandler(values []string, labels []string, index int) (string, error) {
	var builder strings.Builder
	builder.WriteString("index:")
	builder.WriteString(strconv.Itoa(index))
	for i, value := range values {
		if i < len(labels) {
			builder.WriteString("\t")
			builder.WriteString(labels[i])
			builder.WriteString(":")
			if value == "\"-\"" || value == "" {
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
	e, err := json.Marshal(m.Errors)
	if err != nil {
		return "", fmt.Errorf("cannot marshal errors as json: %w", err)
	}
	source := m.Source
	if source == "" {
		source = "-"
	}
	return fmt.Sprintf(
		"total:%d\tmatched:%d\tunmatched:%d\tskipped:%d\tsource:%s\terrors:%s",
		m.Total, m.Matched, m.Unmatched, m.Skipped, source, e,
	), nil
}
