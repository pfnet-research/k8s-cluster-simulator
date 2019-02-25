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
	// Immediately returns ErrEmptyQueue if the queue is empty.
	Pop() (*v1.Pod, error)

	// PlaceBack pushes the pod to the "head" of this queue.
	PlaceBack(pod *v1.Pod)
}
