package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSONLineHandler converts each log line to a JSON (NDJSON) format.
// It serializes the matched parts of the log line into a JSON object,
// with the log line index and matched values as key-value pairs.
func JSONLineHandler(pairs map[string]string, order []string, index int, hasIndex bool) (string, error) {
	om, err := orderMap(pairs, order, index, hasIndex)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(om)
	if err != nil {
		return "", fmt.Errorf("cannot marshal as json: %w", err)
	}
	return string(b), nil
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
func PrettyJSONLineHandler(pairs map[string]string, order []string, index int, hasIndex bool) (string, error) {
	orderMap(pairs, order, index, hasIndex)
	om := make(map[string]string)
	for _, key := range order {
		if value, ok := pairs[key]; ok {
			om[key] = value
		}
	}

	b, err := json.MarshalIndent(orderedData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot marshal as json: %w", err)
	}

	return string(b), nil
}

// PrettyJSONMetadataHandler is similar to JSONMetadataHandler but formats the output in a human-readable JSON format.
// It first converts the metadata to a JSON string and then formats it with indents for better readability.
func PrettyJSONMetadataHandler(m *Metadata) (string, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot marshal result as json: %w", err)
	}
	return string(b), nil
}

// KeyValuePairLineHandler converts log lines into a key-value pair format.
// Each matched log line is represented as a string of key-value pairs, separated by spaces.
func KeyValuePairLineHandler(pairs map[string]string, order []string, index int, hasIndex bool) (string, error) {
	om, err := orderMap(pairs, order, index, hasIndex)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, key := range order {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(key)
		builder.WriteString("=\"")
		builder.WriteString(om[key])
		builder.WriteString("\"")
	}
	return builder.String(), nil
}

// KeyValuePairMetadataHandler converts the metadata into a string of key-value pairs.
// It serializes metadata like total line count, matched count, etc., into a single string.
func KeyValuePairMetadataHandler(m *Metadata) (string, error) {
	var builder strings.Builder
	builder.WriteString("total:\"")
	builder.WriteString(strconv.Itoa(m.Total))
	builder.WriteString("\" matched:\"")
	builder.WriteString(strconv.Itoa(m.Matched))
	builder.WriteString("\" unmatched:\"")
	builder.WriteString(strconv.Itoa(m.Unmatched))
	builder.WriteString("\" skipped:\"")
	builder.WriteString(strconv.Itoa(m.Skipped))
	builder.WriteString("\" source:\"")
	builder.WriteString(m.Source)
	if len(m.Errors) > 0 {
		b, err := json.Marshal(m.Errors)
		if err != nil {
			return "", fmt.Errorf("cannot marshal errors as json: %w", err)
		}
		builder.WriteString("\" errors:\"")
		builder.Write(b)
		builder.WriteString("\"")
	}
	return builder.String(), nil
}

// LTSVLineHandler formats each log line as LTSV (Labeled Tab-separated Values).
// It converts each matched log line into a tab-separated string, with each field as a label-value pair.
func LTSVLineHandler(pairs map[string]string, order []string, index int, hasIndex bool) (string, error) {
	om, err := orderMap(pairs, order, index, hasIndex)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, key := range order {
		if builder.Len() > 0 {
			builder.WriteString("\t")
		}
		builder.WriteString(key)
		builder.WriteString(":")
		builder.WriteString(om[key])
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
	if len(m.Errors) > 0 {
		b, err := json.Marshal(m.Errors)
		if err != nil {
			return "", fmt.Errorf("cannot marshal errors as json: %w", err)
		}
		builder.WriteString("\terrors:")
		builder.Write(b)
	}
	return builder.String(), nil
}

func orderMap(pairs map[string]string, order []string, index int, hasIndex bool) {
	if hasIndex {
		i := strconv.Itoa(index)
		order = append([]string{"index"}, order...)
		pairs["index"] = i
	}
}
