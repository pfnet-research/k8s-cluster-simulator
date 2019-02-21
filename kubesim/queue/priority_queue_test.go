package queue

import (
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPodWithPriority(name string, prio *int32, ts metav1.Time) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: ts,
		},
		Spec: v1.PodSpec{
			Priority: prio,
		},
	}

	return &pod
}

func (pq *PriorityQueue) isSorted(comparator Compare) bool {
	pod, _ := pq.Pop()

	for {
		podNext, err := pq.Pop()
		if err != nil {
			break
		}

		if !comparator(pod, podNext) {
			return false
		}
		pod = podNext
	}

	return true
}

func TestPriorityQueuePushAndPop(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	q.Push(newPodWithPriority("pod-0", nil, now))

	prio := int32(1)
	q.Push(newPodWithPriority("pod-1", &prio, now))

	dur, _ := time.ParseDuration("1s")
	clock := metav1.NewTime(now.Add(dur))
	q.Push(newPodWithPriority("pod-2", &prio, clock))

	pod, _ := q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}
}

func TestPriorityQueuePlaceBack(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	q.PlaceBack(newPodWithPriority("pod-0", nil, now))
	pod, _ := q.Pop()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	q.PlaceBack(newPodWithPriority("pod-1", nil, now))

	prio := int32(1)
	q.PlaceBack(newPodWithPriority("pod-2", &prio, now))

	dur, _ := time.ParseDuration("1s")
	clock := metav1.NewTime(now.Add(dur))
	q.PlaceBack(newPodWithPriority("pod-3", &prio, clock))

	pod, _ = q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
	}
	pod, _ = q.Pop()
	if pod.Name != "pod-3" {
		t.Errorf("got: %v\nwant: \"pod-3\"", pod.Name)
	}
	pod, _ = q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}
}

func TestPriorityQueueIsSorted(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	for prio := 9; prio >= 0; prio-- {
		p := int32(prio)
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &p, now))
	}

	if !q.isSorted(q.inner.comparator) {
		t.Error("PriorityQueue is not sorted")
	}
}

func TestPriorityQueuePendingPods(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	podsNum := 256

	for prio := 0; prio < podsNum; prio++ {
		p := int32(prio)
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &p, now))
	}

	pods := q.PendingPods()
	if len(pods) != podsNum {
		t.Errorf("got: %v\nwant: \"%v\"", len(pods), podsNum)
	}
}

func TestPriorityQueueIsSortedWithCustomComparator(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueueWithComparator(lowPriority)

	for prio := 9; prio >= 0; prio-- {
		p := int32(prio)
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &p, now))
	}

	if !q.isSorted(q.inner.comparator) {
		t.Error("PriorityQueue is not sorted")
	}
}

func lowPriority(pod0, pod1 *v1.Pod) bool {
	prio0 := getPodPriority(pod0)
	prio1 := getPodPriority(pod1)
	return prio0 < prio1
}
