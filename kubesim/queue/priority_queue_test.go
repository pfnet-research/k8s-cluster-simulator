package queue

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPodWithPriority(name string, prio *int32) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Priority: prio,
		},
	}

	return &pod
}

func (pq *rawPriorityQueue) isSorted() bool {
	items := pq.items()
	for i := 1; i < pq.Len(); i++ {
		if getPriority(items[i-1].pod) < getPriority(items[i].pod) {
			return false
		}
	}
	return true
}

func TestPriorityQueuePushAndPop(t *testing.T) {
	q := NewPriorityQueue()

	q.Push(newPodWithPriority("pod-0", nil))
	prio := int32(1)
	q.Push(newPodWithPriority("pod-1", &prio))
	prio = 2
	q.Push(newPodWithPriority("pod-2", &prio))

	pod, _ := q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}
}

func TestPriorityQueuePlaceBack(t *testing.T) {
	q := NewPriorityQueue()

	q.PlaceBack(newPodWithPriority("pod-0", nil))
	pod, _ := q.Pop()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	prio := int32(1)
	q.PlaceBack(newPodWithPriority("pod-1", &prio))
	prio = 2
	q.PlaceBack(newPodWithPriority("pod-2", &prio))

	pod, _ = q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
	}
	pod, _ = q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}
}

func TestPriorityQueueIsSorted(t *testing.T) {
	q := NewPriorityQueue()

	for prio := int32(9); prio >= 0; prio-- {
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &prio))
	}

	if !q.inner.isSorted() {
		t.Error("PriorityQueue is not sorted")
	}

	for _, pod := range q.PendingPods() {
		fmt.Println(pod.Name)
	}
}
