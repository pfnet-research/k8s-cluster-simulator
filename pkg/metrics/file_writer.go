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
	"strings"
)

// FileWriter is a Writer that writes metrics to a file.
type FileWriter struct {
	file      *os.File
	formatter Formatter
}

// NewFileWriter creates a new FileWriter with an output device or file at the given path, and the formatter that
// formats metrics to a string
// If /dev/stdout or stdout is given, the standard out is set.
// If /dev/stderr or stderr is given, the standard error is set.
// Otherwise, the file of a given path is set and it will be truncated if it exists.
// Returns error if failed to create a file.
func NewFileWriter(dest string, formatter Formatter) (*FileWriter, error) {
	var file *os.File
	if dest == "/dev/stdout" || strings.ToLower(dest) == "stdout" {
		file = os.Stdout
	} else if dest == "/dev/stderr" || strings.ToLower(dest) == "stderr" {
		file = os.Stderr
	} else {
		f, err := os.Create(dest)
		if err != nil {
			return nil, err
		}
		file = f
	}

	return &FileWriter{
		file:      file,
		formatter: formatter,
	}, nil
}

// FileName returns the name of file underlying this FileWriter.
func (w *FileWriter) FileName() string { return w.file.Name() }

// Write implements Writer interface.
// Returns error if failed to format with the underlying formatter.
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
