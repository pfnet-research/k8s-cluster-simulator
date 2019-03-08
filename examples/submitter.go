package main

import (
	"fmt"
	"math/rand"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/ordovicia/kubernetes-simulator/api"
	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
)

type mySubmitter struct {
	podIdx        uint64
	targetPodsNum int
}

func newMySubmitter(targetPodsNum int) *mySubmitter {
	rand.Seed(time.Now().UnixNano())

	return &mySubmitter{
		podIdx:        0,
		targetPodsNum: targetPodsNum,
	}
}

func (s *mySubmitter) Submit(
	clock clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]api.SubmitterEvent, error) {

	queueMetrics := met[metrics.QueueMetricsKey].(queue.Metrics)
	submissionNum := s.targetPodsNum - queueMetrics.PendingPodsNum
	if submissionNum <= 0 {
		return []api.SubmitterEvent{}, nil
	}

	events := make([]api.SubmitterEvent, 0, submissionNum+1)

	if s.podIdx > 0 { // Test deleting previously submitted pod
		podName := fmt.Sprintf("pod-%d", s.podIdx-1)
		events = append(events, &api.DeleteEvent{PodNamespace: "default", PodName: podName})
	}

	for i := 0; i < submissionNum; i++ {
		events = append(events, &api.SubmitEvent{Pod: newPod(s.podIdx)})
		s.podIdx++
	}

	return events, nil
}

func newPod(idx uint64) *v1.Pod {
	simSpec := ""
	for i := 0; i < rand.Intn(4)+1; i++ {
		sec := rand.Intn(60*60 + 1)
		cpu := rand.Intn(4 + 1)
		mem := rand.Intn(4 + 1)
		gpu := rand.Intn(1 + 1)

		simSpec += fmt.Sprintf(`
- seconds: %d
  resourceUsage:
    cpu: %d
    memory: %dGi
    nvidia.com/gpu: %d
`, sec, cpu, mem, gpu)
	}

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pod-%d", idx),
			Namespace: "default",
			Annotations: map[string]string{
				"simSpec": simSpec,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "container",
					Image: "container",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse("4"),
							"memory":         resource.MustParse("4Gi"),
							"nvidia.com/gpu": resource.MustParse("1"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("6"),
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
