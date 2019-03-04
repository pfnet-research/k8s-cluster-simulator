package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
)

type mySubmitter struct {
	startClock    clock.Clock
	submissionCnt uint64
}

func (s *mySubmitter) Submit(clock clock.Clock, _ algorithm.NodeLister, _ metrics.Metrics) ([]*v1.Pod, error) {
	if s.submissionCnt == 0 {
		s.startClock = clock
	}

	pods := []*v1.Pod{}
	elapsedSec := clock.Sub(s.startClock).Seconds()

	for s.submissionCnt <= uint64(elapsedSec)/21 {
		pods = append(pods, newPod(s.submissionCnt, clock))
		s.submissionCnt++
	}

	return pods, nil
}

func newPod(n uint64, clock clock.Clock) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pod-%d", n),
			Namespace: "default",
			Annotations: map[string]string{
				"simSpec": `
- seconds: 60
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 90
  resourceUsage:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
`,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "container",
					Image: "container",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse("3"),
							"memory":         resource.MustParse("5Gi"),
							"nvidia.com/gpu": resource.MustParse("1"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("4"),
							"memory":         resource.MustParse("6Gi"),
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	return &pod
}
