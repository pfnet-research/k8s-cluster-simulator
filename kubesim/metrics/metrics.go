package metrics

import (
	"encoding/json"
	"os"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
)

// NodesMetrics is a map associating node names and node.Metrics,
// plus "Type" to nodesMetricsType and "Clock" to a formatted clock.
type NodesMetrics map[string]interface{}

// PodsMetrics is a map associating pod names and pod.Metrics,
// plus "Type" to podsMetricsType and "Clock" to a formatted clock.
type PodsMetrics map[string]interface{}

const (
	nodesMetricsType = "NodesMetrics"
	podsMetricsType  = "PodsMetrics"
)

// BuildMetrics builds NodesMetrics and PodsMetrics.
func BuildMetrics(clock clock.Clock, nodes map[string]*node.Node) (NodesMetrics, PodsMetrics) {
	nodesMetrics := make(map[string]interface{})
	podsMetrics := make(map[string]interface{})

	nodesMetrics["Clock"] = clock.ToRFC3339()
	nodesMetrics["Type"] = nodesMetricsType

	podsMetrics["Clock"] = clock.ToRFC3339()
	podsMetrics["Type"] = podsMetricsType

	for name, node := range nodes {
		nodesMetrics[name] = node.Metrics(clock)
		for _, pod := range node.PodList() {
			if !pod.IsTerminated(clock) {
				podsMetrics[pod.ToV1().Name] = pod.Metrics(clock)
			}
		}
	}

	return nodesMetrics, podsMetrics
}

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
