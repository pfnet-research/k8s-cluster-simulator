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

package scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
)

// Scheduler defines the lowest-level scheduler interface.
type Scheduler interface {
	// Schedule makes scheduling decisions for (subset of) pending pods and running pods.
	// The return value is a list of scheduling events.
	// This method must never block.
	Schedule(
		clock clock.Clock,
		podQueue queue.PodQueue,
		nodeLister algorithm.NodeLister,
		nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error)
}

// Event defines the interface of a scheduling event.
// Submit can returns any type in a list that implements this interface.
type Event interface {
	IsSchedulerEvent() bool
}

// BindEvent represents an event of deciding the binding of a pod to a node.
type BindEvent struct {
	Pod            *v1.Pod
	ScheduleResult core.ScheduleResult
}

// DeleteEvent represents an event of the deleting a bound pod on a node.
type DeleteEvent struct {
	PodNamespace string
	PodName      string
	NodeName     string
}

func (b *BindEvent) IsSchedulerEvent() bool   { return true }
func (d *DeleteEvent) IsSchedulerEvent() bool { return true }
