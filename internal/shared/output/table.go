package output

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/openmarkers/openmarkers-cli/internal/shared/ui"
)

type TableFormatter struct{}

func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

func (f *TableFormatter) Output(data any, columns []Column) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			fmt.Fprintln(os.Stdout, "No results.")
			return nil
		}
		return f.printTable(v, columns)
	}

	tf := NewTextFormatter(false)
	return tf.Output(data, columns)
}

func (f *TableFormatter) Error(code string, message string) error {
	fmt.Fprintf(os.Stderr, "%s %s\n", ui.ErrorStyle.Render("Error ["+code+"]:"), message)
	return nil
}

func (f *TableFormatter) printTable(v reflect.Value, columns []Column) error {
	if len(columns) == 0 {
		tf := NewTextFormatter(false)
		return tf.Output(v.Interface(), columns)
	}

	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = max(len(col.Title), col.Width)
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

	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(os.Stdout, "  ")
		}
		fmt.Fprint(os.Stdout, ui.TableHeader.Width(widths[i]).Render(col.Title))
	}
	fmt.Fprintln(os.Stdout)

	for i, w := range widths {
		if i > 0 {
			fmt.Fprint(os.Stdout, "  ")
		}
		fmt.Fprint(os.Stdout, ui.TableSep.Render(strings.Repeat("─", w)))
	}
	fmt.Fprintln(os.Stdout)

	for _, row := range rows {
		for j, cell := range row {
			if j > 0 {
				fmt.Fprint(os.Stdout, "  ")
			}
			fmt.Fprint(os.Stdout, ui.TableCell.Width(widths[j]).Render(cell))
		}
		fmt.Fprintln(os.Stdout)
	}

	return nil
}
