package metrics

import (
	"encoding/json"
	"os"
)

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

func (f *FileWriter) Write(nodesMetrics NodesMetrics, podsMetrics PodsMetrics) error {
	bytes, err := json.Marshal(nodesMetrics)
	if err != nil {
		return err
	}
	f.file.Write(bytes)
	f.file.Write([]byte{'\n'})

	bytes, err = json.Marshal(podsMetrics)
	if err != nil {
		return err
	}
	f.file.Write(bytes)
	f.file.Write([]byte{'\n'})

	return nil
}

func (f *FileWriter) Close() error {
	return f.file.Close()
}

var _ = Writer(&FileWriter{})
