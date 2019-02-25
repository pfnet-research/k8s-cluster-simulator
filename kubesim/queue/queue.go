package queue

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// PodProducer deifnes the interface of pod producer.
type PodProducer interface {
	// Produce produces all or a portion of the pending pods.
	// This method consumes the pending pods; the same pod will not returned by this method again.
	Produce() ([]*v1.Pod, error)
}

// ErrEmptyQueue is returned from Pop.
var ErrEmptyQueue = errors.New("No pod queued")

// PodQueue defines the interface of pod queues.
type PodQueue interface {
	PodProducer

	// Push pushes the pod to the "end" of this queue.
	Push(pod *v1.Pod)

	// Pop pops a single pod from this queue.
	// Immediately returns ErrEmptyQueue if the queue is empty.
	Pop() (*v1.Pod, error)
}
