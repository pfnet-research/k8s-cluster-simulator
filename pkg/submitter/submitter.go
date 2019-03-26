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

package submitter

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
)

// Submitter defines the submitter interface.
type Submitter interface {
	// Submit submits pods to a simulated cluster.
	// The return value is a list of submitter events.
	// Submitters are called serially in the same order that they are registered to the simulated
	// cluster.
	// This method must never block.
	Submit(
		clock clock.Clock,
		nodeLister algorithm.NodeLister,
		metrics metrics.Metrics) ([]Event, error)
}

// Event defines the interface of a submitter event.
// Submit can returns any type in a list that implements this interface.
type Event interface {
	IsSubmitterEvent() bool
}

// SubmitEvent represents an event of submitting a pod to a cluster.
type SubmitEvent struct {
	Pod *v1.Pod
}

// DeleteEvent represents an event of deleting a pod from a cluster.
type DeleteEvent struct {
	PodName      string
	PodNamespace string
}

// UpdateEvent represents an event of updating the manifest of a pending pod.
type UpdateEvent struct {
	PodName      string
	PodNamespace string
	NewPod       *v1.Pod
}

// TerminateSubmitterEvent represents an event of terminating the submission process.
type TerminateSubmitterEvent struct {
}

func (s *SubmitEvent) IsSubmitterEvent() bool             { return true }
func (d *DeleteEvent) IsSubmitterEvent() bool             { return true }
func (u *UpdateEvent) IsSubmitterEvent() bool             { return true }
func (t *TerminateSubmitterEvent) IsSubmitterEvent() bool { return true }
