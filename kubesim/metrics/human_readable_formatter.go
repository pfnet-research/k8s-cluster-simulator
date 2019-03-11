package metrics

import (
	"fmt"

	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
)

// HumanReadableFormatter formats metrics in a human-readable style.
type HumanReadableFormatter struct{}

func (h *HumanReadableFormatter) Format(metrics Metrics) (string, error) {
	// Clock
	clk, ok := metrics[ClockKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", ClockKey)
	}
	c, ok := clk.(string)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not string", ClockKey, clk)
	}
	str := "Metrics " + c + "\n"

	// Nodes
	str += "  Nodes\n"

	nodesMetrics, ok := metrics[NodesMetricsKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", NodesMetricsKey)
	}
	nodesMet, ok := nodesMetrics.(map[string]node.Metrics)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not map[string]node.Metrics", NodesMetricsKey, nodesMetrics)
	}

	s, err := formatNodesMetrics(nodesMet)
	if err != nil {
		return "", err
	}
	str += s

	// Pods
	str += "  Pods\n"

	podsMetrics, ok := metrics[PodsMetricsKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", PodsMetricsKey)
	}
	podsMet, ok := podsMetrics.(map[string]pod.Metrics)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not map[string]pod.Metrics", PodsMetricsKey, podsMetrics)
	}

	s, err = formatPodsMetrics(podsMet)
	if err != nil {
		return "", err
	}
	str += s

	// Queue
	str += "  Queue\n"

	queueMetrics, ok := metrics[QueueMetricsKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", QueueMetricsKey)
	}
	queueMet, ok := queueMetrics.(queue.Metrics)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not queue.Metrics", QueueMetricsKey, queueMetrics)
	}

	s, err = formatQueueMetrics(queueMet)
	if err != nil {
		return "", err
	}
	str += s

	return str, nil
}

func formatNodesMetrics(metrics map[string]node.Metrics) (string, error) {
	str := ""

	for name, met := range metrics {
		str += fmt.Sprintf("    %s: Pods %d(%d)/%d", name, met.RunningPodsNum, met.TerminatingPodsNum, met.Capacity.Pods().Value())
		for rsrc, cap := range met.Capacity {
			if rsrc == "pods" {
				continue
			}

			usage := met.TotalResourceUsage[rsrc] // !ok -> usage == 0
			req := met.TotalResourceRequest[rsrc]

			if rsrc == "memory" {
				d := int64(1 << 20)
				str += fmt.Sprintf(", memMB %d/%d/%d", usage.Value()/d, req.Value()/d, cap.Value()/d)
			} else {
				str += fmt.Sprintf(", %s %d/%d/%d", rsrc, usage.Value(), req.Value(), cap.Value())
			}
		}

		str += fmt.Sprintf(", Failed %d\n", met.FailedPodsNum)
	}

	return str, nil
}

func formatPodsMetrics(metrics map[string]pod.Metrics) (string, error) {
	str := ""

	for name, met := range metrics {
		str += fmt.Sprintf("    %s: prio %d, bound at %s on %s, status %s, elapsed %d s",
			name, met.Priority, met.BoundAt.ToRFC3339(), met.Node, met.Status, met.ExecutedSeconds)

		for rsrc, req := range met.ResourceRequest {
			lim := met.ResourceLimit[rsrc] // !ok -> usage == 0
			usage := met.ResourceUsage[rsrc]

			if rsrc == "memory" {
				d := int64(1 << 20)
				str += fmt.Sprintf(", memMB %d/%d/%d", usage.Value()/d, req.Value()/d, lim.Value()/d)
			} else {
				str += fmt.Sprintf(", %s %d/%d/%d", rsrc, usage.Value(), req.Value(), lim.Value())
			}
		}

		str += "\n"
	}

	return str, nil
}

func formatQueueMetrics(metrics queue.Metrics) (string, error) {
	return fmt.Sprintf("    PendingPods %d\n", metrics.PendingPodsNum), nil
}

var _ = Formatter(&HumanReadableFormatter{})
