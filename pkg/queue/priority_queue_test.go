// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
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

func (pq *PriorityQueue) isSorted(num int) bool {
	pod, err := pq.Pop()
	if err != nil {
		if err == ErrEmptyQueue {
			return num == 0
		}
		return false
	}
	num--

	for {
		podNext, err := pq.Pop()
		if err != nil {
			if err == ErrEmptyQueue {
				break
			}
			return false
		}
		num--

		if !pq.inner.comparator(pod, podNext) {
			return false
		}
		pod = podNext
	}

	return num == 0
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
	expected := "pod-1"
	if pod.Name != "pod-1" {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}

	pod, _ = q.Pop()
	expected = "pod-2"
	if pod.Name != expected {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}

	pod, _ = q.Pop()
	expected = "pod-0"
	if pod.Name != expected {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}
}

func TestPriorityQueueIsSorted(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	for prio := 9; prio >= 0; prio-- {
		p := int32(prio)
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &p, now))
	}

	if !q.isSorted(10) {
		t.Error("PriorityQueue is not sorted")
	}
}

func TestPriorityQueueIsSortedWithCustomComparator(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueueWithComparator(lowPriority)

	for prio := 9; prio >= 0; prio-- {
		p := int32(prio)
		q.Push(newPodWithPriority(fmt.Sprintf("pod-%d", prio), &p, now))
	}

	if !q.isSorted(10) {
		t.Error("PriorityQueue is not sorted")
	}
}

func lowPriority(pod0, pod1 *v1.Pod) bool {
	prio0 := util.PodPriority(pod0)
	prio1 := util.PodPriority(pod1)
	return prio0 <= prio1
}

func TestPriorityQueueFront(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	q.Push(newPodWithPriority("pod-0", nil, now))

	prio := int32(1)
	q.Push(newPodWithPriority("pod-1", &prio, now))

	dur, _ := time.ParseDuration("1s")
	clock := metav1.NewTime(now.Add(dur))
	q.Push(newPodWithPriority("pod-2", &prio, clock))

	pod, _ := q.Front()
	expected := "pod-1"
	if pod.Name != expected {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}

	pod, _ = q.Front()
	expected = "pod-1"
	if pod.Name != expected {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}

	_, _ = q.Pop()
	pod, _ = q.Front()
	expected = "pod-2"
	if pod.Name != expected {
		t.Errorf("got: %+v\nwant: %+v", pod.Name, expected)
	}

	_, _ = q.Pop()
	_, _ = q.Pop()
	_, err := q.Front()
	if err != ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, ErrEmptyQueue)
	}
}

func TestPriorityReorder(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	q.Push(newPodWithPriority("pod-0", nil, now))

	prio := int32(1)
	q.Push(newPodWithPriority("pod-1", &prio, now))

	dur, _ := time.ParseDuration("1s")
	clock := metav1.NewTime(now.Add(dur))
	q.Push(newPodWithPriority("pod-2", &prio, clock))

	qNew := q.Reorder(lowPriority)
	if !qNew.isSorted(3) {
		t.Error("PriorityQueue is not sorted")
	}
}

func TestPriorityQueueDelete(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	q.Push(newPodWithPriority("pod-0", nil, now))
	q.Push(newPodWithPriority("pod-1", nil, now))

	if !q.Delete("default", "pod-0") {
		t.Errorf("got: false\nwant: true")
	}

	if q.Delete("default", "pod-0") {
		t.Errorf("got: true\nwant: false")
	}

	if !q.Delete("default", "pod-1") {
		t.Errorf("got: false\nwant: true")
	}

	_, err := q.Pop()
	if err != ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, ErrEmptyQueue)
	}
}

func TestPriorityQueueDeleteAndFront(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	prio0 := int32(0)
	q.Push(newPodWithPriority("pod-0", &prio0, now))
	q.Delete("default", "pod-0")

	_, err := q.Front()
	if err != ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, ErrEmptyQueue)
	}
}

func TestPriorityQueueUpdate(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	prio0 := int32(0)
	pod0 := newPodWithPriority("pod-0", &prio0, now)

	err := q.Update("default", "pod-0", pod0)
	assert.EqualError(t, err, "No pod with key \"default/pod-0\"")

	q.Push(pod0)

	pod1 := newPodWithPriority("pod-1", &prio0, now)
	err = q.Update("default", "pod-0", pod1)
	assert.EqualError(t, err, "Original and new pods have different names")

	pod02 := pod0.DeepCopy()
	prio1 := int32(1)
	pod02.Spec.Priority = &prio1
	err = q.Update("default", "pod-0", pod02)
	if err != nil {
		t.Errorf("error %+v", err)
	}

	pod, _ := q.Pop()
	actual := pod.Spec.Priority
	expected := prio1
	if *actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestPriorityQueueNomination(t *testing.T) {
	now := metav1.Now()
	q := NewPriorityQueue()

	pod0 := newPodWithPriority("pod-0", nil, now)

	q.Push(pod0)

	_ = q.UpdateNominatedNode(pod0, "node-0")
	pods := q.NominatedPods("node-0")
	if len(pods) != 1 || pods[0].Name != "pod-0" {
		t.Errorf("got: %v\nwant: [\"pod-0\"]", pods)
	}
}
