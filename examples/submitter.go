package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

// Submitter
type mySubmitter struct {
	startClock clock.Clock
	n          uint64
}

func (s *mySubmitter) Submit(clock clock.Clock, nodes []*v1.Node) (pods []*v1.Pod, err error) {
	if s.n == 0 {
		s.startClock = clock
	}

	elapsed := clock.Sub(s.startClock).Seconds()
	if uint64(elapsed)/5 >= s.n {
		pod := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              fmt.Sprintf("pod-%d", s.n),
				Namespace:         "default",
				CreationTimestamp: clock.ToMetaV1(),
				Annotations: map[string]string{
					"simSpec": `
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10
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

		s.n++
		return []*v1.Pod{&pod}, nil
	}

	return []*v1.Pod{}, nil
}
