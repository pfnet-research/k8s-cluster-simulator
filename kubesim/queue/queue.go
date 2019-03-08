package queue

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// Metrics represents a metrics of a queue.
type Metrics struct {
	PendingPodsNum int
}

// ErrEmptyQueue is returned from Pop.
var ErrEmptyQueue = errors.New("No pod queued")

// PodQueue defines the interface of pod queues.
type PodQueue interface {
	// Push pushes the pod to the "end" of this queue.
	Push(pod *v1.Pod) error

	// Pop pops the pod on the "front" of this queue. Immediately returns ErrEmptyQueue if the
	// queue is empty.
	Pop() (*v1.Pod, error)

	// Front refers (not pops) the pod on the "front" of this queue. Immediately returns
	// ErrEmptyQueue if the queue is empty.
	Front() (*v1.Pod, error)

	// Delete deletes the pod by name from this queue. Returns true if the pod is found, or false
	// otherwise.
	Delete(podNamespace, podName string) (bool, error)

	// Metrics returns a metrics of this queue.
	Metrics() Metrics
}
