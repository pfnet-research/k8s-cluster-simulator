package queue

import (
	v1 "k8s.io/api/core/v1"

	"github.com/ordovicia/k8s-cluster-simulator/kubesim/util"
)

// FIFOQueue stores pods in a FIFO queue.
type FIFOQueue struct {
	// Push adds a pod to both the map and the queue; Pop deletes a pod from both.
	// OTOH, Delete deletes a pod only from the map, so a pod associated with a key popped from the
	// queue may have been deleted.

	pods  map[string]*v1.Pod
	queue []string
}

// NewFIFOQueue creates a new FIFOQueue.
func NewFIFOQueue() *FIFOQueue {
	return &FIFOQueue{
		pods:  map[string]*v1.Pod{},
		queue: []string{},
	}
}

func (fifo *FIFOQueue) Push(pod *v1.Pod) error {
	key, err := util.PodKey(pod)
	if err != nil {
		return err
	}

	fifo.pods[key] = pod
	fifo.queue = append(fifo.queue, key)

	return nil
}

func (fifo *FIFOQueue) Pop() (*v1.Pod, error) {
	for len(fifo.queue) > 0 {
		var key string
		key, fifo.queue = fifo.queue[0], fifo.queue[1:]
		if pod, ok := fifo.pods[key]; ok {
			delete(fifo.pods, key)
			return pod, nil
		}
	}

			return nil, ErrEmptyQueue
		}

func (fifo *FIFOQueue) Front() (*v1.Pod, error) {
	for len(fifo.queue) > 0 {
		key := fifo.queue[0]
		if pod, ok := fifo.pods[key]; ok {
			return pod, nil
		}
		fifo.queue = fifo.queue[1:]
	}

	return nil, ErrEmptyQueue
}

func (fifo *FIFOQueue) Delete(podNamespace, podName string) (bool, error) {
	key := util.PodKeyFromNames(podNamespace, podName)

	_, ok := fifo.pods[key]
	delete(fifo.pods, key)

	return ok, nil
}

func (fifo *FIFOQueue) Update(podNamespace, podName string, newPod *v1.Pod) (bool, error) {
	keyOrig := util.PodKeyFromNames(podNamespace, podName)
	keyNew, err := util.PodKey(newPod)
	if err != nil {
		return false, err
	}
	if keyOrig != keyNew {
		return false, ErrDifferentNames
	}

	_, ok := fifo.pods[keyOrig]
	fifo.pods[keyOrig] = newPod

	return ok, nil
}

// UpdateNominatedNode does nothing. FIFOQueue does not support preemption.
func (fifo *FIFOQueue) UpdateNominatedNode(pod *v1.Pod, nodeName string) error {
	return nil
}

// RemoveNominatedNode does nothing. FIFOQueue does not support preemption.
func (fifo *FIFOQueue) RemoveNominatedNode(pod *v1.Pod) error {
	return nil
}

// NominatedPods does nothing. FIFOQueue does not support preemption.
func (fifo *FIFOQueue) NominatedPods(nodeName string) []*v1.Pod {
	return []*v1.Pod{}
}

func (fifo *FIFOQueue) Metrics() Metrics {
	return Metrics{
		PendingPodsNum: len(fifo.queue),
	}
}

var _ = PodQueue(&FIFOQueue{})
