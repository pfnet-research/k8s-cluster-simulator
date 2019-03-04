package metrics

import (
	"errors"
	"fmt"

	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
)

// HumanReadableFormatter formats metrics in a human-readable style.
type HumanReadableFormatter struct{}

func (h *HumanReadableFormatter) FormatNodesMetrics(nodesMetrics NodesMetrics) (string, error) {
	clk, ok := nodesMetrics["Clock"]
	if !ok {
		return "", errors.New("No \"Clock\" field in nodesMetrics")
	}
	c, ok := clk.(string)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: \"Clock\" field %v is not string", c)
	}

	str := "NodesMetrics " + c + "\n"

	for nodeName, met := range nodesMetrics {
		if nodeName == "Clock" || nodeName == "Type" {
			continue
		}

		m, ok := met.(node.Metrics)
		if !ok {
			return "", fmt.Errorf("Type assertion failed: %v is not node.Metrics", met)
		}

		str += fmt.Sprintf("  %s: Pods %d/%d", nodeName, m.RunningPodsNum, m.Capacity.Pods().Value())
		for rsrc, cap := range m.Capacity {
			if rsrc == "pods" {
				continue
			}

			usage := m.TotalResourceUsage[rsrc] // !ok -> usage == 0
			req := m.TotalResourceRequest[rsrc]

			if rsrc == "memory" {
				d := int64(1 << 20)
				str += fmt.Sprintf(", memMB %d/%d/%d", usage.Value()/d, req.Value()/d, cap.Value()/d)
			} else {
				str += fmt.Sprintf(", %s %d/%d/%d", rsrc, usage.Value(), req.Value(), cap.Value())
			}
		}

		str += fmt.Sprintf(", Failed %d\n", m.FailedPodsNum)
	}

	return str, nil
}

func (h *HumanReadableFormatter) FormatPodsMetrics(podsMetrics PodsMetrics) (string, error) {
	clk, ok := podsMetrics["Clock"]
	if !ok {
		return "", errors.New("No \"Clock\" field in podsMetrics")
	}
	c, ok := clk.(string)
	if !ok {
		return "", fmt.Errorf("Type assertion failed: \"Clock\" field %v is not string", c)
	}

	str := "PodsMetrics " + c + "\n"

	for podName, met := range podsMetrics {
		if podName == "Clock" || podName == "Type" {
			continue
		}

		m, ok := met.(pod.Metrics)
		if !ok {
			return "", fmt.Errorf("Type assertion failed: %v is not pod.Metrics", met)
		}

		str += fmt.Sprintf("  %s: bound on %s at %s, status %s, elapsed %d s",
			podName, m.BoundAt.ToRFC3339(), m.Node, m.Status, m.ExecutedSeconds)

		for rsrc, req := range m.ResourceRequest {
			lim := m.ResourceLimit[rsrc] // !ok -> usage == 0
			usage := m.ResourceUsage[rsrc]

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
