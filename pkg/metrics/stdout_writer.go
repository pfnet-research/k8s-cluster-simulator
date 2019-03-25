package metrics

import (
	"fmt"
)

// StdoutWriter writes metrics to stdout.
type StdoutWriter struct {
	formatter Formatter
}

// NewStdoutWriter creates a new StdoutWriter instance with the formatter.
func NewStdoutWriter(formatter Formatter) StdoutWriter {
	return StdoutWriter{
		formatter: formatter,
	}
}

func (w *StdoutWriter) Write(metrics *Metrics) error {
	str, err := w.formatter.Format(metrics)
	if err != nil {
		return err
	}
	fmt.Println(str)

	return nil
}

var _ = Writer(&StdoutWriter{})
