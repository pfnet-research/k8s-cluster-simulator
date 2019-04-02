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
	"container/heap"

	v1 "k8s.io/api/core/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// PriorityQueue stores pods in a priority queue.
// The pods are sorted by their priority, which can be configured by users.
type PriorityQueue struct {
	// PriorityQueue wraps rawPriorityQueue for type-safetiness.

	inner         rawPriorityQueue
	nominatedPods map[string]map[string]*v1.Pod
}

// Compare is a comparator function that returns true if pod0 has higher priority than pod1, or
// false otherwise.
type Compare = func(pod0, pod1 *v1.Pod) bool

// NewPriorityQueue creates a new PriorityQueue with DefaultComparator.
func NewPriorityQueue() *PriorityQueue {
	return NewPriorityQueueWithComparator(DefaultComparator)
}

// NewPriorityQueueWithComparator creates a new PriorityQueue with the given comparator.
func NewPriorityQueueWithComparator(comparator Compare) *PriorityQueue {
	return newWithItems(map[string]*item{}, comparator)
}

// Reorder creates a new PriorityQueue with all pods stored in the original queue in the sorted
// order according to the given comparator.
func (pq *PriorityQueue) Reorder(comparator Compare) *PriorityQueue {
	pods := pq.inner.pendingPods()
	items := make(map[string]*item, len(pods))
	for idx, pod := range pods {
		key, _ := util.PodKey(pod) // stored pod never have invalid key
		items[key] = &item{pod, idx}
	}

	return newWithItems(items, comparator)
}

func (pq *PriorityQueue) Push(pod *v1.Pod) error {
	if _, err := util.PodKey(pod); err != nil {
		return err
	}

	heap.Push(&pq.inner, &item{pod: pod})
	return nil
}

func (pq *PriorityQueue) Pop() (*v1.Pod, error) {
	if pq.inner.Len() == 0 {
		return nil, ErrEmptyQueue
	}
	return heap.Pop(&pq.inner).(*item).pod, nil
}

func (pq *PriorityQueue) Front() (*v1.Pod, error) {
	if pq.inner.Len() == 0 {
		return nil, ErrEmptyQueue
	}
	return pq.inner.items[pq.inner.keys[0]].pod, nil
}

func (pq *PriorityQueue) Delete(podNamespace, podName string) bool {
	key := util.PodKeyFromNames(podNamespace, podName)
	item, ok := pq.inner.items[key]
	if ok {
		nominatedNodeName := item.pod.Status.NominatedNodeName
		item.pod.Status.NominatedNodeName = ""
		delete(pq.nominatedPods[nominatedNodeName], key)

		heap.Remove(&pq.inner, item.index) // Don't swap
		delete(pq.inner.items, key)        // 	these two lines
	}

	return ok
}

func (pq *PriorityQueue) Update(podNamespace, podName string, newPod *v1.Pod) error {
	keyOrig := util.PodKeyFromNames(podNamespace, podName)
	keyNew, err := util.PodKey(newPod)
	if err != nil {
		return err
	}
	if keyOrig != keyNew {
		return ErrDifferentNames
	}

	if _, ok := pq.inner.items[keyOrig]; !ok {
		return &ErrNoMatchingPod{key: keyOrig}
	}

	pq.inner.items[keyOrig].pod = newPod
	heap.Fix(&pq.inner, pq.inner.items[keyOrig].index)

	return nil
}

func (pq *PriorityQueue) UpdateNominatedNode(pod *v1.Pod, nodeName string) error {
	if err := pq.RemoveNominatedNode(pod); err != nil {
		return err
	}

	pod.Status.NominatedNodeName = nodeName
	key, err := util.PodKey(pod)
	if err != nil {
		return err
	}

	if _, ok := pq.nominatedPods[nodeName]; !ok {
		pq.nominatedPods[nodeName] = map[string]*v1.Pod{}
	}
	pq.nominatedPods[nodeName][key] = pod

	return nil
}

func (pq *PriorityQueue) RemoveNominatedNode(pod *v1.Pod) error {
	nodeName := pod.Status.NominatedNodeName
	if nodeName == "" {
		return nil
	}

	key, err := util.PodKey(pod)
	if err != nil {
		return err
	}

	pod.Status.NominatedNodeName = ""
	delete(pq.nominatedPods[nodeName], key)

	return nil
}

func (pq *PriorityQueue) NominatedPods(nodeName string) []*v1.Pod {
	pods := make([]*v1.Pod, 0, len(pq.nominatedPods[nodeName]))
	for _, pod := range pq.nominatedPods[nodeName] {
		pods = append(pods, pod)
	}

	return pods
}

func (pq *PriorityQueue) Metrics() Metrics {
	return Metrics{
		PendingPodsNum: pq.inner.Len(),
	}
}

var _ = PodQueue(&PriorityQueue{})

type item struct {
	pod   *v1.Pod
	index int // Needed by update and is maintained by the heap.Interface methods.
}

type rawPriorityQueue struct {
	// A pod exists in keys iff it also exists in items.

	items      map[string]*item
	keys       []string
	comparator Compare
}

// Len, Less, and Swap are required to implement sort.Interface, which is included in heap.Interface.
func (pq rawPriorityQueue) Len() int { return len(pq.keys) }

func (pq rawPriorityQueue) Less(i, j int) bool {
	pod0 := pq.items[pq.keys[i]].pod
	pod1 := pq.items[pq.keys[j]].pod

	return (pq.comparator)(pod0, pod1)
}

func (pq rawPriorityQueue) Swap(i, j int) {
	pq.keys[i], pq.keys[j] = pq.keys[j], pq.keys[i]

	pq.items[pq.keys[i]].index = i
	pq.items[pq.keys[j]].index = j
}

// Push and Pop are required to implement heap.Interface.
func (pq *rawPriorityQueue) Push(itm interface{}) {
	item := itm.(*item)
	item.index = len(pq.items)

	key, _ := util.PodKey(item.pod) // Error check is performed in PriorityQueue.Push
	pq.items[key] = item
	pq.keys = append(pq.keys, key)
}

func (pq *rawPriorityQueue) Pop() interface{} {
	keysOld := pq.keys
	n := len(keysOld)

	key := keysOld[n-1]
	item := pq.items[key]
	item.index = -1 // for safety

	delete(pq.items, key)
	pq.keys = keysOld[0 : n-1]

	return item
}

// nolint
func (pq *rawPriorityQueue) front() *item {
	return pq.items[pq.keys[0]]
}

func (pq *rawPriorityQueue) pendingPods() []*v1.Pod {
	pods := make([]*v1.Pod, 0, pq.Len())
	for _, item := range pq.items {
		pods = append(pods, item.pod)
	}
	return pods
}

// DefaultComparator returns true if pod0 has higher priority than pod1.
// If the priorities are equal, it compares the timestamps and returns true if pod0 is older than
// pod1.
func DefaultComparator(pod0, pod1 *v1.Pod) bool {
	prio0 := util.PodPriority(pod0)
	prio1 := util.PodPriority(pod1)

	ts0 := podTimestamp(pod0)
	ts1 := podTimestamp(pod1)

	return (prio0 > prio1) || (prio0 == prio1 && ts0.Before(ts1))
}

func newWithItems(items map[string]*item, comparator Compare) *PriorityQueue {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	rawPq := rawPriorityQueue{
		items:      items,
		keys:       keys,
		comparator: comparator,
	}
	heap.Init(&rawPq)

	return &PriorityQueue{
		inner:         rawPq,
		nominatedPods: map[string]map[string]*v1.Pod{},
	}
}
