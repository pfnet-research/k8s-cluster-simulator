package kubesim

import (
	"errors"
	"sync"

	v1 "k8s.io/api/core/v1"
)

// podQueue stores pods in a queue.
// It wraps []*v1.Pod for thread-safetiness.
type podQueue struct {
	queue []*v1.Pod
	lock  sync.Mutex
}

// append pushes a pod to this podQueue.
func (q *podQueue) append(pod *v1.Pod) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queue = append(q.queue, pod)
}

// errEmptyPodQueue is returned from pop.
var errEmptyPodQueue = errors.New("No pod queued")

// pop pops a pod from this podQueue.
// If this podQueue is empty, errEmptyPodQueue will be returned.
func (q *podQueue) pop() (*v1.Pod, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.queue) == 0 {
		return nil, errEmptyPodQueue
	}

	var pod *v1.Pod
	pod, q.queue = q.queue[0], q.queue[1:]

	return pod, nil
}
