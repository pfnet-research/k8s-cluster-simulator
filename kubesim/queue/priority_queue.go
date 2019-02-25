package queue

import (
	"container/heap"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/scheduling"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

// PriorityQueue stores pods in a priority queue.
// The pods are sorted by their priority.
//
// PriorityQueue wraps rawPriorityQueue for type-safetiness.
type PriorityQueue struct {
	inner rawPriorityQueue
}

// Compare returns true if pod0 has higher priority than pod1.
// Otherwise, this function returns false.
type Compare = func(pod0, pod1 *v1.Pod) bool

// NewPriorityQueue creates a new PriorityQueue with DefaultComparator.
func NewPriorityQueue() *PriorityQueue {
	return NewPriorityQueueWithComparator(DefaultComparator)
}

// NewPriorityQueueWithComparator creates a new PriorityQueue with the given comparator function.
func NewPriorityQueueWithComparator(comparator Compare) *PriorityQueue {
	rawPq := rawPriorityQueue{
		items:      make([]*item, 0),
		comparator: comparator,
	}
	heap.Init(&rawPq)

	return &PriorityQueue{
		inner: rawPq,
	}
}

// Reorder creates a new PriorityQueue. All pods stored in the original PriorityQueue are moved to
// the new one, in the sorted order according to the given comparator.
func (pq *PriorityQueue) Reorder(comparator Compare) *PriorityQueue {
	pods := pq.inner.pendingPods()
	pqNew := NewPriorityQueueWithComparator(comparator)
	for _, pod := range pods {
		pqNew.Push(pod)
	}

	return pqNew
}

func (pq *PriorityQueue) Push(pod *v1.Pod) {
	heap.Push(&pq.inner, &item{pod: pod})
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
	return pq.inner.items[0].pod, nil
}

var _ = PodQueue(&PriorityQueue{}) // Making sure that PriorityQueue implements PodQueue.

type item struct {
	pod   *v1.Pod
	index int // Needed by update and is maintained by the heap.Interface methods.
}

type rawPriorityQueue struct {
	items      []*item
	comparator Compare
}

// Len, Less, and Swap are required to implement sort.Interface, which is included in heap.Interface.
func (pq rawPriorityQueue) Len() int { return len(pq.items) }

func (pq rawPriorityQueue) Less(i, j int) bool {
	pod0 := pq.items[i].pod
	pod1 := pq.items[j].pod

	return (pq.comparator)(pod0, pod1)
}

func (pq rawPriorityQueue) Swap(i, j int) {
	items := pq.items
	items[i], items[j] = items[j], items[i]
	items[i].index = i
	items[j].index = j
}

// Push and Pop are required to implement heap.Interface.
func (pq *rawPriorityQueue) Push(itm interface{}) {
	item := itm.(*item)
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *rawPriorityQueue) Pop() interface{} {
	pqOld := pq.items
	n := len(pqOld)
	item := pqOld[n-1]
	item.index = -1 // for safety
	pq.items = pqOld[0 : n-1]

	return item
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
	prio0 := getPodPriority(pod0)
	prio1 := getPodPriority(pod1)

	ts0 := getPodTimestamp(pod0)
	ts1 := getPodTimestamp(pod1)

	return (prio0 > prio1) || (prio0 == prio1 && ts0.Before(ts1))
}

func getPodPriority(pod *v1.Pod) int32 {
	prio := int32(scheduling.DefaultPriorityWhenNoDefaultClassExists)
	if pod.Spec.Priority != nil {
		prio = *pod.Spec.Priority
	}
	return prio
}

func getPodTimestamp(pod *v1.Pod) clock.Clock {
	// _, condition := podutil.GetPodCondition(&pod.Status, v1.PodScheduled)
	// if condition == nil {
	ts := clock.NewClockWithMetaV1(pod.CreationTimestamp)
	return ts
	// }
	// if condition.LastProbeTime.IsZero() {
	// 	return &condition.LastTransitionTime
	// }
	// return &condition.LastProbeTime
}
