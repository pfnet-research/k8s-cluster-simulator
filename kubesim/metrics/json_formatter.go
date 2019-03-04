package metrics

import (
	"encoding/json"
)

// JSONFormatter formats metrics to a JSON string.
type JSONFormatter struct {
}

// FormatNodesMetrics formats the given nodesMetrics to a JSON string, without newline at the end.
func (j *JSONFormatter) FormatNodesMetrics(nodesMetrics NodesMetrics) (string, error) {
	bytes, err := json.Marshal(nodesMetrics)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// FormatPodsMetrics formats the given podsMetrics to a JSON string, without newline at the end.
func (j *JSONFormatter) FormatPodsMetrics(podsMetrics PodsMetrics) (string, error) {
	bytes, err := json.Marshal(podsMetrics)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

var _ = Formatter(&JSONFormatter{})
