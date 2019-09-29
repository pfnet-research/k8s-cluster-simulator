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
	"io/ioutil"
	"math/rand"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
)

type mySubmitter struct {
	podIdx       uint64
	totalPodsNum uint64
	myrand       *rand.Rand
}

func newMySubmitter(totalPodsNum uint64) *mySubmitter {
	return &mySubmitter{
		podIdx:       0,
		totalPodsNum: totalPodsNum,
		myrand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *mySubmitter) generateWorkloads(
	clock clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {

	// delete the workload folder first.
	// os.RemoveAll(workloadPath)
	// err := os.MkdirAll(workloadPath, 755)
	// if err != nil {
	// 	log.L.Fatalf("Cannot create the folder %v", workloadPath)
	// 	return nil, fmt.Errorf("Cannot create the folder %v", workloadPath)
	// }

	// find the number of submission at clock time.
	submissionNum := int(s.myrand.Int31n(3))
	events := make([]submitter.Event, 0, submissionNum+1)

	for i := 0; i < submissionNum; i++ {
		events = append(events, &submitter.SubmitEvent{Pod: s.newRandomPod(s.podIdx, clock)})
		s.podIdx++
		if s.podIdx >= uint64(s.totalPodsNum) {
			events = append(events, &submitter.TerminateSubmitterEvent{})
			return events, nil
		}
	}

	return events, nil
}

func (s *mySubmitter) Submit(
	clock clock.Clock,
	_ algorithm.NodeLister,
	met metrics.Metrics) ([]submitter.Event, error) {
	return s.generateWorkloads(clock, nil, met)
}

func (s *mySubmitter) newPod(idx uint64, prio int32, phaseNum int, secs []uint64,
	cpuUsages []uint64, memUsages []uint64, gpuUsages []uint64,
	cpuRequest uint64, memRequest uint64, gpuRequest uint64,
	cpuLimit uint64, memLimit uint64, gpuLimit uint64) (*v1.Pod, error) {

	if (phaseNum != len(cpuUsages)) || (phaseNum != len(memUsages)) || (phaseNum != len(gpuUsages)) {
		return nil, fmt.Errorf("phaseNum mismatch: %d", phaseNum, len(cpuUsages), len(memUsages), len(gpuUsages))
	}

	simSpec := ""
	for i := 0; i < int(phaseNum); i++ {
		simSpec += fmt.Sprintf(`
- seconds: %d
  resourceUsage:
    cpu: %d
    memory: %dGi
    nvidia.com/gpu: %d
`, secs[i], cpuUsages[i], memUsages[i], gpuUsages[i])
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
			Priority: &prio,
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: "container",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse(fmt.Sprintf("%d", cpuRequest)),
							"memory":         resource.MustParse(fmt.Sprintf("%dGi", memRequest)),
							"nvidia.com/gpu": resource.MustParse(fmt.Sprintf("%d", gpuRequest)),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse(fmt.Sprintf("%d", cpuLimit)),
							"memory":         resource.MustParse(fmt.Sprintf("%dGi", memLimit)),
							"nvidia.com/gpu": resource.MustParse(fmt.Sprintf("%d", gpuLimit)),
						},
					},
				},
			},
		},
	}

	return &pod, nil
}

func (s *mySubmitter) loadPod(fileName string) (*v1.Pod, error) {
	return nil, nil
}

func (s *mySubmitter) newRandomPod(idx uint64, clock clock.Clock) *v1.Pod {
	simSpec := ""

	for i := 0; i < s.myrand.Intn(4)+1; i++ {
		sec := 60 * s.myrand.Intn(60)
		cpu := 1 + s.myrand.Intn(4)
		mem := 0
		gpu := 0

		simSpec += fmt.Sprintf(`
- seconds: %d
  resourceUsage:
    cpu: %d
    memory: %dGi
    nvidia.com/gpu: %d
`, sec, cpu, mem, gpu)
	}
	// prio := s.myrand.Int31n(3) / 2 // 0, 0, 1
	prio := int32(1) // TODO(tanle): nil memory if we set all pods'pirority as the same priority

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

	d, err := yaml.Marshal(&pod)
	if err != nil {
		log.L.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s-pod%d.yaml", workloadPath, clock, idx), d, 0644)
	if err != nil {
		log.L.Fatal(err)
	}

	return &pod
}
