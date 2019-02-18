package queue

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// PodQueue stores pods in a queue.
type PodQueue struct {
	queue []*v1.Pod
}

// Append pushes a pod to this PodQueue.
func (q *PodQueue) Append(pod *v1.Pod) {
	q.queue = append(q.queue, pod)
}

// ErrEmptyPodQueue is returned from Pop.
var ErrEmptyPodQueue = errors.New("No pod queued")

// Pop pops a pod from this PodQueue.
// If this PodQueue is empty, errEmptyPodQueue will be returned.
func (q *PodQueue) Pop() (*v1.Pod, error) {
	if len(q.queue) == 0 {
		return nil, ErrEmptyPodQueue
	}

	var pod *v1.Pod
	pod, q.queue = q.queue[0], q.queue[1:]

	return pod, nil
}

// PlaceBack pushes a pod to the head of this PodQueue.
func (q *PodQueue) PlaceBack(pod *v1.Pod) {
	q.queue = append([]*v1.Pod{pod}, q.queue...)
}
