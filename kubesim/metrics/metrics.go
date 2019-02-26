package metrics

// NodesMetrics is a map associating node names and node.Metrics, plus "clock" to a formatted clock.
type NodesMetrics map[string]interface{}

// PodsMetrics is a map associating pod names and pod.Metrics, plus "clock" to a formatted clock.
type PodsMetrics map[string]interface{}

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given NodesMetrics and PodsMetrics to some location.
	Write(nodeMetrics NodesMetrics, podsMetrics PodsMetrics) error
}
