package queue

import (
	"github.com/ordovicia/kubernetes-simulator/kubesim/scheduler"
	v1 "k8s.io/api/core/v1"
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

// Produce implements the PodProducer interface.
// This methods returns all pending pods.
func (fifo *FIFOQueue) Produce() ([]*v1.Pod, error) {
	return fifo.q, nil
}

var _ = scheduler.PodProducer(&FIFOQueue{})
