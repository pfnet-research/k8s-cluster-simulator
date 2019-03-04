package queue

import (
	v1 "k8s.io/api/core/v1"
)

// FIFOQueue stores pods in a FIFO queue.
type FIFOQueue struct {
	q []*v1.Pod
}

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

func (fifo *FIFOQueue) Front() (*v1.Pod, error) {
	if len(fifo.q) == 0 {
		return nil, ErrEmptyQueue
	}
	return fifo.q[0], nil
}

func (fifo *FIFOQueue) Metrics() Metrics {
	return Metrics{
		PendingPodsNum: len(fifo.q),
	}
}

var _ = PodQueue(&FIFOQueue{})
