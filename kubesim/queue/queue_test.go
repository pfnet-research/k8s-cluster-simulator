package queue

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPod(name string) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
	return &pod
}

func TestFIFOQueuePush(t *testing.T) {
	q := FIFOQueue{}

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
}

func TestFIFOPlaceBack(t *testing.T) {
	q := FIFOQueue{}

	q.PlaceBack(newPod("pod-0"))
	pod, _ := q.Pop()
	if pod.Name != "pod-0" {
		t.Errorf("got: %v\nwant: \"pod-0\"", pod.Name)
	}

	q.PlaceBack(newPod("pod-1"))
	q.PlaceBack(newPod("pod-2"))
	pod, _ = q.Pop()
	if pod.Name != "pod-2" {
		t.Errorf("got: %v\nwant: \"pod-2\"", pod.Name)
	}
	pod, _ = q.Pop()
	if pod.Name != "pod-1" {
		t.Errorf("got: %v\nwant: \"pod-1\"", pod.Name)
	}
}
