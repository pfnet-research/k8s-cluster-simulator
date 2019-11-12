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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
)

type mySubmitter struct {
	podIdx       uint64
	totalPodsNum uint64
	myrand       *rand.Rand
	tick         time.Duration
	endClock     clock.Clock
}

func newMySubmitter(totalPodsNum uint64, endClock clock.Clock) *mySubmitter {
	return &mySubmitter{
		podIdx:       0,
		totalPodsNum: totalPodsNum,
		myrand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		tick:         time.Duration(10), //TODO: get tick from viper.config
		endClock:     endClock,
	}
}

func (s *mySubmitter) generateWorkloads(
	clock clock.Clock,
	met metrics.Metrics) ([]submitter.Event, error) {

	queueMetrics := met[metrics.QueueMetricsKey].(queue.Metrics)
	nodesMetrics := met[metrics.NodesMetricsKey].(map[string]node.Metrics)
	runningPodsNum := int(0)
	for _, met := range nodesMetrics {
		runningPodsNum += int(met.RunningPodsNum)
	}
	submissionNum := targetNum - queueMetrics.PendingPodsNum - runningPodsNum
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

func (s *mySubmitter) loadWorkload(
	clock clock.Clock,
	met metrics.Metrics) ([]submitter.Event, error) {
	// find all files matching with clocks.
	// startClock := clock
	// endClock := clock.Add(s.tick)
	podNames := podMap[clock.ToRFC3339()]
	// load the pods and submit them.

	events := make([]submitter.Event, 0, len(podNames)+1)

	for _, podName := range podNames {
		fileName := fmt.Sprintf("%s@%s", clock.ToRFC3339(), podName)
		filePath := fmt.Sprintf("%s/%s", workloadPath, fileName)
		newPod, err := s.loadPod(filePath)
		if err != nil {
			log.L.Errorf("cannot load %s", filePath)
			return events, fmt.Errorf("cannot load %s", filePath)
		}
		submittedPodsNum++
		events = append(events, &submitter.SubmitEvent{Pod: newPod})

		if submittedPodsNum >= uint64(s.totalPodsNum) {
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
	// terminate the simulation before end time

	if s.endClock.Before(clock) {
		log.L.Infof("=========== Terminate simumation ========== @ %v", clock.ToRFC3339())
		events := make([]submitter.Event, 0, 1)
		events = append(events, &submitter.TerminateSubmitterEvent{})
		return events, nil
	}

	if isGenWorkload {
		return s.generateWorkloads(clock, met)
	}

	return s.loadWorkload(clock, met)
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

func (s *mySubmitter) loadPod(filePath string) (*v1.Pod, error) {
	d, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.L.Errorf("cannot find file %s", filePath)
		return nil, fmt.Errorf("cannot find file %s", filePath)
	}

	pod := v1.Pod{}

	// err = yaml.Unmarshal(d, &pod)
	// if err != nil {
	// 	log.L.Errorf("cannot parse pod from file %s", filePath)
	// 	return nil, fmt.Errorf("cannot parse pod from file %s", filePath)
	// }

	err = json.Unmarshal(d, &pod)
	if err != nil {
		log.L.Errorf("cannot parse pod from file %s", filePath)
		return nil, fmt.Errorf("cannot parse pod from file %s", filePath)
	}

	return &pod, nil
}

func (s *mySubmitter) newRandomPod(idx uint64, clock clock.Clock) *v1.Pod {
	simSpec := ""
	for i := 0; i < phasNum; i++ {

		sec := int(genNormFloat64(meanSec/2, meanSec, meanSec*0.1+1, meanSec*10, s.myrand))
		cpu := 1 + int(genNormFloat64(cpuStd, meanCpu, meanCpu*0.1, meanCpu*10, s.myrand))
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
							"cpu":            resource.MustParse(fmt.Sprintf("%d", int(requestCpu))),
							"memory":         resource.MustParse("0Gi"),
							"nvidia.com/gpu": resource.MustParse("0"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("999"),
							"memory":         resource.MustParse("999Gi"),
							"nvidia.com/gpu": resource.MustParse("0"),
						},
					},
				},
			},
		},
	}

	WritePodAsJson(pod, workloadPath, clock)
	return &pod
}
