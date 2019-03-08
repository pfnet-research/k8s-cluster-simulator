package metrics

import (
	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// Metrics represents a metrics at one time point, in the following structure.
//   Metrics[ClockKey] = a formatted clock
//   Metrics[NodesMetricsKey] = map from node name to node.Metrics
//   Metrics[PodsMetricsKey] = map from pod name to pod.Metrics
// 	 Metrics[QueueMetricsKey] = queue.Metrics
type Metrics map[string]interface{}

const (
	// ClockKey is the key associated to clock.
	ClockKey = "Clock"
	// NodesMetricsKey is the key associated to a map of node.Metrics.
	NodesMetricsKey = "Nodes"
	// PodsMetricsKey is the key associated to a map of pod.Metrics.
	PodsMetricsKey = "Pods"
	// QueueMetricsKey is the key associated to queue.Metrics.
	QueueMetricsKey = "Queue"
)

// BuildMetrics builds a Metrics at the time clock.
func BuildMetrics(clock clock.Clock, nodes map[string]*node.Node, queue queue.PodQueue) (Metrics, error) {
	metrics := make(map[string]interface{})
	metrics[ClockKey] = clock.ToRFC3339()

	nodesMetrics := make(map[string]node.Metrics)
	podsMetrics := make(map[string]pod.Metrics)

	for name, node := range nodes {
		nodesMetrics[name] = node.Metrics(clock)
		for _, pod := range node.PodList() {
			if !pod.IsTerminated(clock) {
				key, err := util.PodKey(pod.ToV1())
				if err != nil {
					return Metrics{}, err
				}
				podsMetrics[key] = pod.Metrics(clock)
			}
		}
	}

	metrics[NodesMetricsKey] = nodesMetrics
	metrics[PodsMetricsKey] = podsMetrics
	metrics[QueueMetricsKey] = queue.Metrics()

	return metrics, nil
}

// Formatter defines the interface of metrics formatter.
type Formatter interface {
	// Format formats the given metrics to a string.
	Format(metrics Metrics) (string, error)
}

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given metrics to some location.
	Write(metrics Metrics) error
}
