package metrics

import (
	"fmt"

	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
)

// HumanReadableFormatter formats metrics in a human-readable style.
type HumanReadableFormatter struct{}

func (h *HumanReadableFormatter) Format(metrics Metrics) (string, error) {
	// Clock
	clk, ok := metrics[clockKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", clockKey)
	}
	c, ok := clk.(string)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not string", clockKey, clk)
	}
	str := "Metrics " + c + "\n"

	// Nodes
	str += "  Nodes\n"

	nodesMetrics, ok := metrics[nodesMetricsKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", nodesMetricsKey)
	}
	nodesMet, ok := nodesMetrics.(map[string]node.Metrics)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not map[string]node.Metrics", nodesMetricsKey, nodesMetrics)
	}

	s, err := formatNodesMetrics(nodesMet)
	if err != nil {
		return "", err
	}
	str += s

	// Pods
	str += "  Pods\n"

	podsMetrics, ok := metrics[podsMetricsKey]
	if !ok {
		return "", fmt.Errorf("No %q field in metrics", podsMetricsKey)
	}
	podsMet, ok := podsMetrics.(map[string]pod.Metrics)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: %q field %v is not map[string]pod.Metrics", podsMetricsKey, podsMetrics)
	}

	s, err = formatPodsMetrics(podsMet)
	if err != nil {
		return "", err
	}
	str += s

	return str, nil
}

func formatNodesMetrics(metrics map[string]node.Metrics) (string, error) {
	str := ""

	for name, met := range metrics {
		str += fmt.Sprintf("    %s: Pods %d/%d", name, met.RunningPodsNum, met.Capacity.Pods().Value())
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
		str += fmt.Sprintf("    %s: bound on %s at %s, status %s, elapsed %d s",
			name, met.BoundAt.ToRFC3339(), met.Node, met.Status, met.ExecutedSeconds)

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

var _ = Formatter(&HumanReadableFormatter{})
