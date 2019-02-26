package metrics

import (
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

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given NodesMetrics and PodsMetrics to some location.
	Write(nodeMetrics NodesMetrics, podsMetrics PodsMetrics) error

	// Close this writer.
	Close() error
}

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
