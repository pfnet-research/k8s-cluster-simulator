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
	"fmt"
)

// StdoutWriter is a Writer that writes metrics to stdout.
type StdoutWriter struct {
	formatter Formatter
}

// NewStdoutWriter creates a new StdoutWriter with the formatter.
func NewStdoutWriter(formatter Formatter) StdoutWriter {
	return StdoutWriter{
		formatter: formatter,
	}
}

// Write implements Writer interface.
// Returns error if failed to format with the underlying formatter.
func (w *StdoutWriter) Write(metrics *Metrics) error {
	str, err := w.formatter.Format(metrics)
	if err != nil {
		return err
	}
	fmt.Println(str)

	return nil
}

var _ = Writer(&StdoutWriter{})
