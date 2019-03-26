// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"os"
)

// FileWriter writes metrics to a file.
type FileWriter struct {
	file      *os.File
	formatter Formatter
}

// NewFileWriter creates a new FileWriter instance with a file at the given path, and a formatter
// that formats metrics to a string. Returns err if failed to create a file.
func NewFileWriter(path string, formatter Formatter) (*FileWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:      file,
		formatter: formatter,
	}, nil
}

// FileName returns the name of file underlying this FileWriter.
func (w *FileWriter) FileName() string { return w.file.Name() }

func (w *FileWriter) Write(metrics *Metrics) error {
	str, err := w.formatter.Format(metrics)
	if err != nil {
		return err
	}
	w.file.WriteString(str)
	w.file.Write([]byte{'\n'})

	return nil
}

var _ = Writer(&FileWriter{})
