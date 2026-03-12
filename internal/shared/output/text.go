package output

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type TextFormatter struct {
	NoColor bool
}

func NewTextFormatter(noColor bool) *TextFormatter {
	return &TextFormatter{NoColor: noColor}
}

func (f *TextFormatter) Output(data any, columns []Column) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			fmt.Fprintln(os.Stdout, "No results.")
			return nil
		}
		return f.printSlice(v, columns)
	}

	if v.Kind() == reflect.Map {
		return f.printMap(v)
	}

	if v.Kind() == reflect.String {
		fmt.Fprintln(os.Stdout, v.String())
		return nil
	}

	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		return f.printStruct(v)
	}

	fmt.Fprintf(os.Stdout, "%v\n", data)
	return nil
}

func (f *TextFormatter) Error(code string, message string) error {
	fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", code, message)
	return nil
}

func (f *TextFormatter) printSlice(v reflect.Value, columns []Column) error {
	if len(columns) == 0 {
		for i := 0; i < v.Len(); i++ {
			fmt.Fprintf(os.Stdout, "%v\n", v.Index(i).Interface())
		}
		return nil
	}

	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col.Title)
		if col.Width > widths[i] {
			widths[i] = col.Width
		}
	}

	rows := make([][]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		row := extractRow(v.Index(i), columns)
		rows[i] = row
		for j, cell := range row {
			if len(cell) > widths[j] {
				widths[j] = len(cell)
			}
		}
	}

	header := make([]string, len(columns))
	for i, col := range columns {
		header[i] = padRight(col.Title, widths[i])
	}
	fmt.Fprintln(os.Stdout, strings.Join(header, "  "))

	sep := make([]string, len(columns))
	for i, w := range widths {
		sep[i] = strings.Repeat("─", w)
	}
	fmt.Fprintln(os.Stdout, strings.Join(sep, "  "))

	for _, row := range rows {
		cells := make([]string, len(columns))
		for j, cell := range row {
			cells[j] = padRight(cell, widths[j])
		}
		fmt.Fprintln(os.Stdout, strings.Join(cells, "  "))
	}

	return nil
}

func (f *TextFormatter) printMap(v reflect.Value) error {
	for _, key := range v.MapKeys() {
		fmt.Fprintf(os.Stdout, "%v: %v\n", key.Interface(), v.MapIndex(key).Interface())
	}
	return nil
}

func (f *TextFormatter) printStruct(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	maxKey := 0
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if len(name) > maxKey {
			maxKey = len(name)
		}
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fmt.Fprintf(os.Stdout, "%-*s  %v\n", maxKey, field.Name+":", v.Field(i).Interface())
	}
	return nil
}

func extractRow(v reflect.Value, columns []Column) []string {
	row := make([]string, len(columns))
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		for i, col := range columns {
			val := v.MapIndex(reflect.ValueOf(col.Key))
			if val.IsValid() {
				row[i] = fmt.Sprintf("%v", val.Interface())
			}
		}
	case reflect.Struct:
		t := v.Type()
		for i, col := range columns {
			for j := 0; j < t.NumField(); j++ {
				tag := t.Field(j).Tag.Get("json")
				name := strings.Split(tag, ",")[0]
				if name == col.Key || t.Field(j).Name == col.Key {
					row[i] = fmt.Sprintf("%v", v.Field(j).Interface())
					break
				}
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			return extractRow(v.Elem(), columns)
		}
	}
	return row
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
