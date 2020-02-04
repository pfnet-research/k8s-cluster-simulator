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

package main

import (
	"fmt"
	"math/rand"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
)

type mySubmitter struct {
	podIdx        uint64
	targetPodsNum int
	myrand        *rand.Rand
}

func newMySubmitter(targetPodsNum int) *mySubmitter {
	return &mySubmitter{
		podIdx:        0,
		targetPodsNum: targetPodsNum,
		myrand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *mySubmitter) Submit(
	clock clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {

	queueMetrics := met[metrics.QueueMetricsKey].(queue.Metrics)
	submissionNum := s.targetPodsNum - queueMetrics.PendingPodsNum
	if submissionNum <= 0 {
		return []submitter.Event{}, nil
	}

	events := make([]submitter.Event, 0, submissionNum+1)

	if s.podIdx > 0 { // Test deleting previously submitted pod
		podName := fmt.Sprintf("pod-%d", s.podIdx-1)
		events = append(events, &submitter.DeleteEvent{PodNamespace: "default", PodName: podName})
	}

	for i := 0; i < submissionNum; i++ {
		events = append(events, &submitter.SubmitEvent{Pod: s.newPod(s.podIdx)})
		s.podIdx++
	}

	if s.podIdx > 1024 {
		events = append(events, &submitter.TerminateSubmitterEvent{})
	}

	return events, nil
}

func (s *mySubmitter) newPod(idx uint64) *v1.Pod {
	simSpec := ""
	for i := 0; i < s.myrand.Intn(4)+1; i++ {
		sec := 60 * s.myrand.Intn(60)
		cpu := 1 + s.myrand.Intn(4)
		mem := 1 + s.myrand.Intn(4)
		gpu := s.myrand.Intn(2)

		simSpec += fmt.Sprintf(`
- seconds: %d
  resourceUsage:
    cpu: %d
    memory: %dGi
    nvidia.com/gpu: %d
`, sec, cpu, mem, gpu)
	}

	prio := s.myrand.Int31n(3) / 2 // 0, 0, 1

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
			Priority: &prio,
			Containers: []v1.Container{
				{
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
