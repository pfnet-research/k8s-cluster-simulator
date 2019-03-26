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
	"sort"

	v1 "k8s.io/api/core/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
)

// TableFormatter is a Formatter that formats metrics in a table.
type TableFormatter struct{}

// Format implements Formatter interface.
// Returns error if the given metrics does not have valid structure.
func (t *TableFormatter) Format(metrics *Metrics) (string, error) {
	if err := validateMetrics(metrics); err != nil {
		return "", err
	}

	// Clock
	clk := (*metrics)[ClockKey].(string)
	str := clk + "\n\n"

	// Nodes
	nodesMet := (*metrics)[NodesMetricsKey].(map[string]node.Metrics)
	s, resourceTypes := t.formatNodesMetrics(nodesMet)
	str += s + "\n"

	// Pods
	podsMet := (*metrics)[PodsMetricsKey].(map[string]pod.Metrics)
	str += t.formatPodsMetrics(podsMet, resourceTypes) + "\n"

	// Queue
	queueMet := (*metrics)[QueueMetricsKey].(queue.Metrics)
	str += t.formatQueueMetrics(queueMet) + "\n"

	return str, nil
}

var _ = Formatter(&TableFormatter{})

func (t *TableFormatter) formatNodesMetrics(metrics map[string]node.Metrics) (string, []string) {
	nodes, resourceTypes := t.sortedNodeNamesAndResourceTypes(metrics)

	// Header
	str := "Node             Pods   Termi- Failed Capa-  "
	for _, r := range resourceTypes {
		if r == "memory" {
			str += "memory (MB)                   "
		} else {
			str += fmt.Sprintf("%-29s ", r)
		}
	}
	str += "\n"
	str += "                        nating        city   "
	line := ""
	for range resourceTypes {
		str += "Usage    Request  Allocatable "
		line += "------------------------------"
	}
	str += "\n"
	str += "---------------------------------------------" + line + "\n"

	// Body
	for _, node := range nodes {
		met := metrics[node]

		str += fmt.Sprintf(
			"%-16s %-6d %-6d %-6d %-6d ",
			node, met.RunningPodsNum, met.TerminatingPodsNum, met.FailedPodsNum, met.Allocatable.Pods().Value())

		for _, rsrc := range resourceTypes {
			r := v1.ResourceName(rsrc)
			alloc := met.Allocatable[r]
			req := met.TotalResourceRequest[r]
			usg := met.TotalResourceUsage[r]

			allocatable := alloc.Value()
			requested := req.Value()
			usage := usg.Value()

			if rsrc == "memory" {
				d := int64(1 << 20)
				allocatable /= d
				requested /= d
				usage /= d
			}

			str += fmt.Sprintf("%-8d %-8d %-11d ", usage, requested, allocatable)
		}

		str += "\n"
	}

	return str, resourceTypes
}

func (t *TableFormatter) formatPodsMetrics(metrics map[string]pod.Metrics, resourceTypes []string) string {
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

	return str
}

func (t *TableFormatter) formatQueueMetrics(metrics queue.Metrics) string {
	str := "      PendingPods \n"
	str += "------------------\n"
	str += fmt.Sprintf("Queue %-8d \n", metrics.PendingPodsNum)
	return str
}

func (t *TableFormatter) sortedNodeNamesAndResourceTypes(metrics map[string]node.Metrics) ([]string, []string) {
	nodes := make([]string, 0, len(metrics))

	type void struct{}
	rsrcTypes := map[string]void{}

	for name, met := range metrics {
		nodes = append(nodes, name)
		for rsrc := range met.Allocatable {
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
