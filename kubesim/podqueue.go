package kubesim

import (
	"errors"
	"sync"

	"k8s.io/api/core/v1"
)

type podQueue struct {
	queue []v1.Pod
	lock  sync.Mutex
}

var errNoPod = errors.New("No pod queued")

func (q *podQueue) append(pod v1.Pod) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queue = append(q.queue, pod)
}

// pop pops a pod from this podQueue.
// If this podQueue is empty, errNoPod will be returned.
func (q *podQueue) pop() (*v1.Pod, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.queue) == 0 {
		return nil, errNoPod
	}

	var pod v1.Pod
	pod, q.queue = q.queue[0], q.queue[1:]

	return &pod, nil
}
