package metrics

import (
	"fmt"
	"sort"

	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	v1 "k8s.io/api/core/v1"
)

// HumanReadableFormatter formats metrics in a human-readable style.
type TableFormatter struct{}

func (t *TableFormatter) Format(metrics *Metrics) (string, error) {
	if err := validateMetrics(metrics); err != nil {
		return "", err
	}

	// Clock
	clk := (*metrics)[ClockKey].(string)
	str := clk + "\n\n"

	// Nodes
	nodesMet := (*metrics)[NodesMetricsKey].(map[string]node.Metrics)
	s, resourceTypes, err := t.formatNodesMetrics(nodesMet)
	if err != nil {
		return "", err
	}
	str += s + "\n"

	// Pods
	podsMet := (*metrics)[PodsMetricsKey].(map[string]pod.Metrics)
	s, err = t.formatPodsMetrics(podsMet, resourceTypes)
	if err != nil {
		return "", err
	}
	str += s + "\n"

	// Queue
	queueMet := (*metrics)[QueueMetricsKey].(queue.Metrics)
	s, err = t.formatQueueMetrics(queueMet)
	if err != nil {
		return "", err
	}
	str += s

	return str, nil
}

var _ = Formatter(&TableFormatter{})

func (t *TableFormatter) formatNodesMetrics(metrics map[string]node.Metrics) (string, []string, error) {
	nodes, resourceTypes := t.sortedNodeNamesAndResourceTypes(metrics)

	// Header
	str := "Node             Pods   Termi- Failed Capa-  "
	for _, r := range resourceTypes {
		if r == "memory" {
			str += "memory (MB)                "
		} else {
			str += fmt.Sprintf("%-26s ", r)
		}
	}
	str += "\n"
	str += "                        nating        city   "
	line := ""
	for range resourceTypes {
		str += "Usage    Request  Capacity "
		line += "---------------------------"
	}
	str += "\n"
	str += "---------------------------------------------" + line + "\n"

	// Body
	for _, node := range nodes {
		met := metrics[node]

		str += fmt.Sprintf(
			"%-16s %-6d %-6d %-6d %-6d ",
			node, met.RunningPodsNum, met.TerminatingPodsNum, met.FailedPodsNum, met.Capacity.Pods().Value())

		for _, rsrc := range resourceTypes {
			r := v1.ResourceName(rsrc)
			cap := met.Capacity[r]
			req := met.TotalResourceRequest[r]
			usg := met.TotalResourceUsage[r]

			capacity := cap.Value()
			requested := req.Value()
			usage := usg.Value()

			if rsrc == "memory" {
				d := int64(1 << 20)
				capacity /= d
				requested /= d
				usage /= d
			}

			str += fmt.Sprintf("%-8d %-8d %-8d ", usage, requested, capacity)
		}

		str += "\n"
	}

	return str, resourceTypes, nil
}

func (t *TableFormatter) formatPodsMetrics(metrics map[string]pod.Metrics, resourceTypes []string) (string, error) {
	pods := t.sortedPodNames(metrics)

	// Header
	str := "Pod                  Status       Priority Node     BoundAt                   Executed "
	for _, r := range resourceTypes {
		if r == "memory" {
			str += "memory (MB)                "
		} else {
			str += fmt.Sprintf("%-26s ", r)
		}
	}
	str += "\n"
	str += "                                                                              Seconds  "
	line := ""
	for range resourceTypes {
		str += "Usage    Request  Limit    "
		line += "---------------------------"
	}
	str += "\n"
	str += "---------------------------------------------------------------------------------------" + line + "\n"

	// Body
	for _, pod := range pods {
		met := metrics[pod]

		str += fmt.Sprintf(
			"%-20s %-12s %-8d %-8s %-25s %-8d ",
			pod, met.Status, met.Priority, met.Node, met.BoundAt.ToRFC3339(), met.ExecutedSeconds)

		for _, rsrc := range resourceTypes {
			r := v1.ResourceName(rsrc)
			lim := met.ResourceLimit[r]
			req := met.ResourceRequest[r]
			usg := met.ResourceUsage[r]

			limit := lim.Value()
			requested := req.Value()
			usage := usg.Value()

			if rsrc == "memory" {
				d := int64(1 << 20)
				limit /= d
				requested /= d
				usage /= d
			}

			str += fmt.Sprintf("%-8d %-8d %-8d ", usage, requested, limit)
		}

		str += "\n"
	}

	return str, nil
}

func (t *TableFormatter) formatQueueMetrics(metrics queue.Metrics) (string, error) {
	str := "      PendingPods \n"
	str += "------------------\n"
	str += fmt.Sprintf("Queue %-8d \n", metrics.PendingPodsNum)
	return str, nil
}

func (t *TableFormatter) sortedNodeNamesAndResourceTypes(metrics map[string]node.Metrics) ([]string, []string) {
	nodes := make([]string, 0, len(metrics))

	type void struct{}
	rsrcTypes := map[string]void{}

	for name, met := range metrics {
		nodes = append(nodes, name)
		for rsrc := range met.Capacity {
			rsrcTypes[rsrc.String()] = void{}
		}
	}

	resourceTypes := make([]string, 0, len(rsrcTypes))
	for rsrc := range rsrcTypes {
		if rsrc != "pods" {
			resourceTypes = append(resourceTypes, rsrc)
		}
	}

	sort.Strings(nodes)
	sort.Strings(resourceTypes)

	return nodes, resourceTypes
}

func (t *TableFormatter) sortedPodNames(metrics map[string]pod.Metrics) []string {
	pods := make([]string, 0, len(metrics))
	for name := range metrics {
		pods = append(pods, name)
	}
	sort.Strings(pods)
	return pods
}
