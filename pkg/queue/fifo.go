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
	v1 "k8s.io/api/core/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// FIFOQueue stores pods in a FIFO queue.
type FIFOQueue struct {
	// Push adds a pod to both the map and the slice; Pop deletes a pod from both.
	// OTOH, Delete deletes a pod only from the map, so a pod associated with a key popped from the
	// slice may have been deleted.
	// Pop and Front check whether the pod actually exists in the slice.

	pods  map[string]*v1.Pod
	queue []string
}

// NewFIFOQueue creates a new FIFOQueue.
func NewFIFOQueue() *FIFOQueue {
	return &FIFOQueue{
		pods:  map[string]*v1.Pod{},
		queue: []string{},
	}
}

func (fifo *FIFOQueue) Push(pod *v1.Pod) error {
	key, err := util.PodKey(pod)
	if err != nil {
		return err
	}

	fifo.pods[key] = pod
	fifo.queue = append(fifo.queue, key)

	return nil
}

func (fifo *FIFOQueue) Pop() (*v1.Pod, error) {
	for len(fifo.queue) > 0 {
		var key string
		key, fifo.queue = fifo.queue[0], fifo.queue[1:]
		if pod, ok := fifo.pods[key]; ok {
			delete(fifo.pods, key)
			return pod, nil
		}
	}

	return nil, ErrEmptyQueue
}

func (fifo *FIFOQueue) Front() (*v1.Pod, error) {
	for len(fifo.queue) > 0 {
		key := fifo.queue[0]
		if pod, ok := fifo.pods[key]; ok {
			return pod, nil
		}
		fifo.queue = fifo.queue[1:]
	}

	return nil, ErrEmptyQueue
}

func (fifo *FIFOQueue) Delete(podNamespace, podName string) bool {
	key := util.PodKeyFromNames(podNamespace, podName)
	_, ok := fifo.pods[key]
	delete(fifo.pods, key)

	return ok
}

func (fifo *FIFOQueue) Update(podNamespace, podName string, newPod *v1.Pod) error {
	keyOrig := util.PodKeyFromNames(podNamespace, podName)
	keyNew, err := util.PodKey(newPod)
	if err != nil {
		return err
	}
	if keyOrig != keyNew {
		return ErrDifferentNames
	}

	if _, ok := fifo.pods[keyOrig]; !ok {
		return &ErrNoMatchingPod{key: keyOrig}
	}

	fifo.pods[keyOrig] = newPod
	return nil
}

// UpdateNominatedNode does nothing. FIFOQueue doesn't support preemption.
func (fifo *FIFOQueue) UpdateNominatedNode(pod *v1.Pod, nodeName string) error {
	return nil
}

// RemoveNominatedNode does nothing. FIFOQueue doesn't support preemption.
func (fifo *FIFOQueue) RemoveNominatedNode(pod *v1.Pod) error {
	return nil
}

// NominatedPods does nothing. FIFOQueue doesn't support preemption.
func (fifo *FIFOQueue) NominatedPods(nodeName string) []*v1.Pod {
	return []*v1.Pod{}
}

func (fifo *FIFOQueue) Metrics() Metrics {
	return Metrics{
		PendingPodsNum: len(fifo.queue),
	}
}

var _ = PodQueue(&FIFOQueue{})
