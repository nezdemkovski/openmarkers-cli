package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONFormatter struct {
	meta map[string]any
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) SetMeta(key string, value any) {
	if f.meta == nil {
		f.meta = make(map[string]any)
	}
	f.meta[key] = value
}

func (f *JSONFormatter) Output(data any, _ []Column) error {
	envelope := map[string]any{"data": data}
	if len(f.meta) > 0 {
		envelope["meta"] = f.meta
		f.meta = nil
	}
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
