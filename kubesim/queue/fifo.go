package queue

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
)

// FIFOQueue stores pods in a FIFO queue.
type FIFOQueue struct {
	q []*v1.Pod
}

var _ = Queue(&FIFOQueue{}) // Making sure that FIFOQueue implements Queue.

func (fifo *FIFOQueue) Push(pod *v1.Pod) {
	fifo.q = append(fifo.q, pod)
}

func (fifo *FIFOQueue) Pop() (*v1.Pod, error) {
	if len(fifo.q) == 0 {
		return nil, ErrEmptyQueue
	}

	var pod *v1.Pod
	pod, fifo.q = fifo.q[0], fifo.q[1:]

	return pod, nil
}

func (fifo *FIFOQueue) PlaceBack(pod *v1.Pod) {
	fifo.q = append([]*v1.Pod{pod}, fifo.q...)
}

func (fifo *FIFOQueue) PendingPods() []*v1.Pod {
	return fifo.q
}

// List and FilteredList implement algorithm.PodLister interface.
// FIFOQueue ignores labels.Selecter.
func (fifo *FIFOQueue) List(_ labels.Selector) ([]*v1.Pod, error) {
	pod, err := fifo.Pop()
	return []*v1.Pod{pod}, err
}

// FilteredList and List implement algorithm.PodLister interface.
// FIFOQueue ignores algorithm.PodFilter and labels.Selecter.
func (fifo *FIFOQueue) FilteredList(_ algorithm.PodFilter, _ labels.Selector) ([]*v1.Pod, error) {
	pod, err := fifo.Pop()
	return []*v1.Pod{pod}, err
}

var _ = algorithm.PodLister(&FIFOQueue{})
