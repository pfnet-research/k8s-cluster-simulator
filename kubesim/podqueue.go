package kubesim

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// podQueue stores pods in a queue.
type podQueue struct {
	queue []*v1.Pod
}

// append pushes a pod to this podQueue.
func (q *podQueue) append(pod *v1.Pod) {
	q.queue = append(q.queue, pod)
}

// errEmptyPodQueue is returned from pop.
var errEmptyPodQueue = errors.New("No pod queued")

// pop pops a pod from this podQueue.
// If this podQueue is empty, errEmptyPodQueue will be returned.
func (q *podQueue) pop() (*v1.Pod, error) {
	if len(q.queue) == 0 {
		return nil, errEmptyPodQueue
	}

	var pod *v1.Pod
	pod, q.queue = q.queue[0], q.queue[1:]

	return pod, nil
}

// placeBack pushes a pod to the head of this podQueue.
func (q *podQueue) placeBack(pod *v1.Pod) {
	q.queue, q.queue[0] = append(q.queue[:1], q.queue[0:]...), pod
}
