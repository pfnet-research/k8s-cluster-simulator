package queue

import (
	"container/heap"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// TODO: Update pod
// TODO: Pod timestamp

// PriorityQueue stores pods in a priority queue.
// The pods are sorted by their priority.
//
// PriorityQueue wraps rawPriorityQueue for type-safetiness.
type PriorityQueue struct {
	inner rawPriorityQueue
}

// NewPriorityQueue creates a new PriorityQueue.
func NewPriorityQueue() *PriorityQueue {
	rawPq := make(rawPriorityQueue, 0)
	heap.Init(&rawPq)

	pq := PriorityQueue{inner: rawPq}
	return &pq
}

func getPriority(pod *v1.Pod) int32 {
	prio := int32(scheduling.DefaultPriorityWhenNoDefaultClassExists)
	if pod.Spec.Priority != nil {
		prio = *pod.Spec.Priority
	}
	return prio
}

func (pq *PriorityQueue) Push(pod *v1.Pod) {
	item := item{
		pod:      pod,
		priority: getPriority(pod),
	}
	heap.Push(&pq.inner, &item)
}

func (pq *PriorityQueue) Pop() (*v1.Pod, error) {
	if pq.inner.Len() == 0 {
		return nil, ErrEmptyQueue
	}

	return heap.Pop(&pq.inner).(*item).pod, nil
}

func (pq *PriorityQueue) PlaceBack(pod *v1.Pod) {
	pq.Push(pod)
}

func (pq *PriorityQueue) PendingPods() []*v1.Pod {
	return pq.inner.pendingPods()
}

type item struct {
	pod      *v1.Pod
	priority int32
	index    int
}

type rawPriorityQueue []*item

// Len, Less, and Swap are required to implement sort.Interface, which is included in heap.Interface.
func (pq rawPriorityQueue) Len() int { return len(pq) }

func (pq rawPriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq rawPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push and Pop are required to implement heap.Interface.
func (pq *rawPriorityQueue) Push(itm interface{}) {
	item := itm.(*item)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *rawPriorityQueue) Pop() interface{} {
	pqOld := *pq
	n := len(pqOld)
	item := pqOld[n-1]
	item.index = -1 // for safety
	*pq = pqOld[0 : n-1]

	return item
}

func (pq *rawPriorityQueue) pendingPods() []*v1.Pod {
	pods := make([]*v1.Pod, 0, pq.Len())
	for _, item := range *pq {
		pods = append(pods, item.pod)
	}
	return pods
}

func (pq *rawPriorityQueue) items() []*item {
	items := make([]*item, 0, pq.Len())
	for _, item := range *pq {
		items = append(items, item)
	}
	return items
}
