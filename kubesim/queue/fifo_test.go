package queue_test

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/k8s-cluster-simulator/kubesim/queue"
)

func newPod(name string) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
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
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}

	pod, _ = q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
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
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	pod, _ = q.Front()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	_, _ = q.Pop()
	pod, _ = q.Front()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}

	_, _ = q.Pop()
	_, _ = q.Pop()
	_, err := q.Front()
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %v\nwant: %v", err, queue.ErrEmptyQueue)
	}
}

func TestFIFOQueueDelete(t *testing.T) {
	q := queue.NewFIFOQueue()

	q.Push(newPod("pod-0"))
	q.Push(newPod("pod-1"))

	ok, _ := q.Delete("default", "pod-0")
	if !ok {
		t.Errorf("got: false\nwant: true")
	}

	ok, _ = q.Delete("default", "pod-0")
	if ok {
		t.Errorf("got: true\nwant: false")
	}

	ok, _ = q.Delete("default", "pod-1")
	if !ok {
		t.Errorf("got: false\nwant: true")
	}

	_, err := q.Pop()
	if err != queue.ErrEmptyQueue {
		t.Errorf("got: %+v\nwant: %+v", err, queue.ErrEmptyQueue)
	}
}
