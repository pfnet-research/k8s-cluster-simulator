package metrics

import (
	"encoding/json"
)

// JSONFormatter formats metrics to a JSON string.
type JSONFormatter struct {
}

// Format formats the given metrics to a JSON string, without newline at the end.
func (j *JSONFormatter) Format(metrics Metrics) (string, error) {
	bytes, err := json.Marshal(metrics)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

var _ = Formatter(&JSONFormatter{})
