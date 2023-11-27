package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// DefaultLineHandler is the default handler that converts log line to NDJSON format.
// It is used if no specific LineHandler is provided when the Parser is instantiated.
func DefaultLineHandler(matches []string, fields []string, index int) (string, error) {
	var builder strings.Builder
	builder.WriteString("{\"index\":")
	builder.WriteString(strconv.Itoa(index))
	for i, match := range matches {
		if i < len(fields) {
			builder.WriteString(",\"")
			builder.WriteString(fields[i])
			if match == "\"-\"" {
				builder.WriteString("\":\"-\"")
				continue
			}
			builder.WriteString("\":")
			b, err := json.Marshal(match)
			if err != nil {
				return "", fmt.Errorf("cannot marshal matched string \"%s\" as json: %w", match, err)
			}
			builder.Write(b)
		}
	}
	builder.WriteString("}")
	return builder.String(), nil
}

// DefaultMetadataHandler is the default handler that converts  metadata to NDJSON format.
// It is used if no specific MetadataHandler is provided when the Parser is instantiated.
func DefaultMetadataHandler(m *Metadata) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("cannot marshal result as json: %w", err)
	}
	return string(b), nil
}
