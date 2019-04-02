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

package queue

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// Metrics represents a metrics of a PodQueue at one time point.
type Metrics struct {
	PendingPodsNum int
}

var (
	// ErrEmptyQueue is returned from Pop.
	ErrEmptyQueue = errors.New("No pod queued")
	// ErrDifferentNames is returned from Update.
	ErrDifferentNames = errors.New("Original and new pods have different names")
)

// ErrNoMatchingPod is returned from Update.
type ErrNoMatchingPod struct {
	key string
}

func (e *ErrNoMatchingPod) Error() string {
	return fmt.Sprintf("No pod with key %q", e.key)
}

// PodQueue defines the interface of pod queues.
type PodQueue interface {
	// Push pushes the pod to the "end" of this PodQueue.
	Push(pod *v1.Pod) error

	// Pop pops the pod on the "front" of this PodQueue.
	// This method never blocks; Immediately returns ErrEmptyQueue if the queue is empty.
	Pop() (*v1.Pod, error)

	// Front refers (not pops) the pod on the "front" of this PodQueue.
	// This method never bocks; Immediately returns ErrEmptyQueue if the queue is empty.
	Front() (*v1.Pod, error)

	// Delete deletes the pod from this PodQueue.
	// Returns true if the pod is found, or false otherwise.
	Delete(podNamespace, podName string) bool

	// Update updates the pod to the newPod.
	// Returns ErrNoMatchingPod if an original pod is not found.
	// The original and new pods must have the same namespace/name; Otherwise ErrDifferentNames is
	// returned in the second field.
	Update(podNamespace, podName string, newPod *v1.Pod) error

	// NominatedPods returns a list of pods for which the node is nominated for scheduling.
	NominatedPods(nodeName string) []*v1.Pod

	// UpdateNominatedNode updates the node nomination for the pod.
	UpdateNominatedNode(pod *v1.Pod, nodeName string) error

	// RemoveNominatedNode removes the node nomination for the pod.
	RemoveNominatedNode(pod *v1.Pod) error

	// Metrics returns a metrics of this PodQueue.
	Metrics() Metrics
}
