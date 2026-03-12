package output

import (
	"fmt"
	"io"
	"os"
)

const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitUsage     = 2
	ExitAuth      = 3
	ExitNotFound  = 4
	ExitServerErr = 5
)

type Column struct {
	Title string
	Width int
	Key   string
}

type Formatter interface {
	Output(data any, columns []Column) error
	Error(code string, message string) error
	SetMeta(key string, value any)
}

type Writer struct {
	Out       io.Writer
	ErrOut    io.Writer
	Formatter Formatter
}

func NewWriter(f Formatter) *Writer {
	return &Writer{
		Out:       os.Stdout,
		ErrOut:    os.Stderr,
		Formatter: f,
	}
}

func (w *Writer) Output(data any, columns []Column) error {
	return w.Formatter.Output(data, columns)
}

func (w *Writer) Error(code string, message string) error {
	return w.Formatter.Error(code, message)
}

func (w *Writer) SetMeta(key string, value any) {
	w.Formatter.SetMeta(key, value)
}

func (w *Writer) Verbose(format string, args ...any) {
	fmt.Fprintf(w.ErrOut, format+"\n", args...)
}
