package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Output(data any, _ []Column) error {
	envelope := map[string]any{"data": data}
	return writeJSON(envelope)
}

func (f *JSONFormatter) Error(code string, message string) error {
	envelope := map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	return writeJSON(envelope)
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
