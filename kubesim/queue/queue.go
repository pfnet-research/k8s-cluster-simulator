package queue

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// ErrEmptyQueue is returned from Pop.
var ErrEmptyQueue = errors.New("No pod queued")

// Queue defines the interface of pod queues.
type Queue interface {
	// Push pushes the pod to the "end" of this queue.
	Push(pod *v1.Pod)

	// Pop pops a single pod from this queue.
	// Returns ErrEmptyQueue if the queue is empty.
	Pop() (*v1.Pod, error)

	// PlaceBack pushes the pod to the "head" of this queue.
	PlaceBack(pod *v1.Pod)
}

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

func (fifo *FIFOQueue) PlaceBack(pod *v1.Pod) {
	fifo.q = append([]*v1.Pod{pod}, fifo.q...)
}
