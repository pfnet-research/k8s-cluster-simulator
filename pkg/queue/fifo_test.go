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

package queue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
)

func newPod(name string) *v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
	}
	return &pod
}

func TestFIFOQueuePushAndPop(t *testing.T) {
	q := queue.NewFIFOQueue()

	q.Push(newPod("pod-0"))
	q.Push(newPod("pod-1"))
	q.Push(newPod("pod-2"))

	pod, _ := q.Pop()
	expected := "pod-0"
	if pod.Name != expected {
		t.Errorf("got: %v\nwant: %q\n", pod.Name, expected)
	}

	pod, _ = q.Pop()
	expected = "pod-1"
	if pod.Name != expected {
		t.Errorf("got: %v\nwant: %q\n", pod.Name, expected)
	}

	pod, _ = q.Pop()
	expected = "pod-2"
	if pod.Name != expected {
		t.Errorf("got: %v\nwant: %q\n", pod.Name, expected)
	}

	_, err := q.Pop()
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %v\nwant: %v", err, queue.ErrEmptyQueue)
	}
}

func TestFIFOQueueFront(t *testing.T) {
	q := queue.NewFIFOQueue()

	q.Push(newPod("pod-0"))
	q.Push(newPod("pod-1"))
	q.Push(newPod("pod-2"))

	pod, _ := q.Front()
	expected := "pod-0"
	if pod.Name != expected {
		t.Errorf("got: %q\nwant: %q", pod.Name, expected)
	}

	pod, _ = q.Front()
	if pod.Name != expected {
		t.Errorf("got: %q\nwant: %q", pod.Name, expected)
	}

	_, _ = q.Pop()
	pod, _ = q.Front()
	expected = "pod-1"
	if pod.Name != expected {
		t.Errorf("got: %q\nwant: %q", pod.Name, expected)
	}

	_, _ = q.Pop()
	_, _ = q.Pop()
	_, err := q.Front()
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, queue.ErrEmptyQueue)
	}
}

func TestFIFOQueueDelete(t *testing.T) {
	q := queue.NewFIFOQueue()

	q.Push(newPod("pod-0"))
	q.Push(newPod("pod-1"))

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
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, queue.ErrEmptyQueue)
	}
}

func TestFIFOQueueDeleteAndFront(t *testing.T) {
	q := queue.NewFIFOQueue()

	q.Push(newPod("pod-0"))
	q.Delete("default", "pod-0")

	_, err := q.Front()
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, queue.ErrEmptyQueue)
	}
}

func TestFIFOQueueUpdate(t *testing.T) {
	q := queue.NewFIFOQueue()

	pod0 := newPod("pod-0")

	err := q.Update("default", "pod-0", pod0)
	assert.EqualError(t, err, "No pod with key \"default/pod-0\"")

	q.Push(pod0)

	pod1 := newPod("pod-1")
	err = q.Update("default", "pod-0", pod1)
	assert.EqualError(t, err, "Original and new pods have different names")

	pod02 := pod0.DeepCopy()
	prio := int32(1)
	pod02.Spec.Priority = &prio
	err = q.Update("default", "pod-0", pod02)
	if err != nil {
		t.Errorf("error %+v", err)
	}

	pod, _ := q.Pop()
	actual := pod.Spec.Priority
	expected := prio
	if *actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}
