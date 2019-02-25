package queue

import (
	v1 "k8s.io/api/core/v1"
)

// FIFOQueue stores pods in a FIFO queue.
type FIFOQueue struct {
	q []*v1.Pod
}

// Produce returns all pending pods.
func (fifo *FIFOQueue) Produce() ([]*v1.Pod, error) {
	return fifo.q, nil
}

var _ = PodProducer(&FIFOQueue{}) // Making sure that FIFOQueue implements PodProducer

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

var _ = PodQueue(&FIFOQueue{})
