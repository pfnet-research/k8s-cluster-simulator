// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
)

// HumanReadableFormatter is a Foramtter that formats metrics in a human-readable style.
type HumanReadableFormatter struct{}

// Format implements Formatter interface.
// Returns error if the given metrics does not have valid structure.
func (h *HumanReadableFormatter) Format(metrics *Metrics) (string, error) {
	if err := validateMetrics(metrics); err != nil {
		return "", err
	}

	// Clock
	clk := (*metrics)[ClockKey].(string)
	str := "Metrics " + clk + "\n"

	// Nodes
	str += "  Nodes\n"
	nodesMet := (*metrics)[NodesMetricsKey].(map[string]node.Metrics)
	str += h.formatNodesMetrics(nodesMet)

	// Pods
	str += "  Pods\n"
	podsMet := (*metrics)[PodsMetricsKey].(map[string]pod.Metrics)
	str += h.formatPodsMetrics(podsMet)

	// Queue
	str += "  Queue\n"
	queueMet := (*metrics)[QueueMetricsKey].(queue.Metrics)
	str += h.formatQueueMetrics(queueMet)

	return str, nil
}

func validateMetrics(metrics *Metrics) error {
	keys := []string{ClockKey, NodesMetricsKey, PodsMetricsKey, QueueMetricsKey}
	for _, key := range keys {
		if _, ok := (*metrics)[key]; !ok {
			return fmt.Errorf("No key %q in metrics", key)
		}
	}

	if _, ok := (*metrics)[ClockKey].(string); !ok {
		return fmt.Errorf("Type assertion failed: %q field of metrics is not string", ClockKey)
	}
	if _, ok := (*metrics)[NodesMetricsKey].(map[string]node.Metrics); !ok {
		return fmt.Errorf("Type assertion failed: %q field of metrics is not map[string]node.Metrics", NodesMetricsKey)
	}
	if _, ok := (*metrics)[PodsMetricsKey].(map[string]pod.Metrics); !ok {
		return fmt.Errorf("Type assertion failed: %q field of metrics is not map[string]pod.Metrics", PodsMetricsKey)
	}
	if _, ok := (*metrics)[QueueMetricsKey].(queue.Metrics); !ok {
		return fmt.Errorf("Type assertion failed: %q field of metrics is not queue.Metrics", QueueMetricsKey)
	}

	return nil
}

func (h *HumanReadableFormatter) formatNodesMetrics(metrics map[string]node.Metrics) string {
	str := ""

	for name, met := range metrics {
		str += fmt.Sprintf("    %s: Pods %d(%d)/%d", name, met.RunningPodsNum, met.TerminatingPodsNum, met.Allocatable.Pods().Value())
		for rsrc, alloc := range met.Allocatable {
			if rsrc == "pods" {
				continue
			}

			usage := met.TotalResourceUsage[rsrc]
			req := met.TotalResourceRequest[rsrc]

			if rsrc == "memory" {
				d := int64(1 << 20)
				str += fmt.Sprintf(", memMB %d/%d/%d", usage.Value()/d, req.Value()/d, alloc.Value()/d)
			} else {
				str += fmt.Sprintf(", %s %d/%d/%d", rsrc, usage.Value(), req.Value(), alloc.Value())
			}
		}

		str += fmt.Sprintf(", Failed %d\n", met.FailedPodsNum)
	}

	return str
}

func (h *HumanReadableFormatter) formatPodsMetrics(metrics map[string]pod.Metrics) string {
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

	return str
}

func (h *HumanReadableFormatter) formatQueueMetrics(metrics queue.Metrics) string {
	return fmt.Sprintf("    PendingPods %d\n", metrics.PendingPodsNum)
}

var _ = Formatter(&HumanReadableFormatter{})
