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

func (w *StdoutWriter) Write(nodesMetrics NodesMetrics, podsMetrics PodsMetrics) error {
	nodesStr, err := w.formatter.FormatNodesMetrics(nodesMetrics)
	if err != nil {
		return err
	}
	fmt.Println(nodesStr)

	podsStr, err := w.formatter.FormatPodsMetrics(podsMetrics)
	if err != nil {
		return err
	}
	fmt.Println(podsStr)

	return nil
}

var _ = Writer(&StdoutWriter{})
