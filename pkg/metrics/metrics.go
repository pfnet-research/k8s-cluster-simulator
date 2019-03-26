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
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// Metrics represents a metrics at one time point, in the following structure.
//   Metrics[ClockKey] = a formatted clock
//   Metrics[NodesMetricsKey] = map from node name to node.Metrics
//   Metrics[PodsMetricsKey] = map from pod name to pod.Metrics
// 	 Metrics[QueueMetricsKey] = queue.Metrics
type Metrics map[string]interface{}

const (
	// ClockKey is the key associated to a clock.Clock.
	ClockKey = "Clock"
	// NodesMetricsKey is the key associated to a map of node.Metrics.
	NodesMetricsKey = "Nodes"
	// PodsMetricsKey is the key associated to a map of pod.Metrics.
	PodsMetricsKey = "Pods"
	// QueueMetricsKey is the key associated to a queue.Metrics.
	QueueMetricsKey = "Queue"
)

// BuildMetrics builds a Metrics at the given clock.
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
	Format(metrics *Metrics) (string, error)
}

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given metrics to some location(s).
	Write(metrics *Metrics) error
}
