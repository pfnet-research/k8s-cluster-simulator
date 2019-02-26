package metrics

import (
	"encoding/json"
	"os"
	"unsafe"
)

// NodesMetrics is a map associating node names and node.Metrics, plus "clock" to a formatted clock.
type NodesMetrics map[string]interface{}

// PodsMetrics is a map associating pod names and pod.Metrics, plus "clock" to a formatted clock.
type PodsMetrics map[string]interface{}

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given NodesMetrics and PodsMetrics to some location.
	Write(nodeMetrics NodesMetrics, podsMetrics PodsMetrics) error
}

// FileWriter writes metrics to a file.
type FileWriter struct {
	file *os.File
}

// NewFileWriter creates a new FileWriter instance with a file at the given path. Returns err if
// failed to create a file.
func NewFileWriter(path string) (*FileWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &FileWriter{file}, nil
}

func (w *FileWriter) Write(nodesMetrics NodesMetrics, podsMetrics PodsMetrics) error {
	bytes, err := json.Marshal(nodesMetrics)
	if err != nil {
		return err
	}
	w.file.Write(bytes)
	w.file.Write([]byte{'\n'})

	bytes, err = json.Marshal(podsMetrics)
	if err != nil {
		return err
	}
	w.file.Write(bytes)
	w.file.Write([]byte{'\n'})

	return nil
}

var _ = Writer(&FileWriter{})
