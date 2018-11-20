// Copyright © 2017 The virtual-kubelet authors
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

// Modification copyright @ 2018 <Name> <E-mail>

package sim

import (
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-scheduler-simulator/log"
)

// Node represents a simulated computing node.
type Node struct {
	config        NodeConfig
	pods          podMap
	resourceUsage v1.ResourceList
}

type NodeConfig struct {
	Name     string
	Capacity v1.ResourceList
	Labels   map[string]string
	Taints   []v1.Taint
}

// NewNode creates a new node with the provided config.
func NewNode(config NodeConfig) Node {
	return Node{
		config:        config,
		pods:          podMap{},
		resourceUsage: v1.ResourceList{},
	}
}

func (node *Node) update(clock Time) {
	node.resourceUsage = v1.ResourceList{}

	node.pods.foreach(func(key string, pod simPod) bool {
		if pod.isTerminated(clock) {
			// pod.status = simPodTerminated
		} else {
			node.resourceUsage = sumResourceList(node.resourceUsage, pod.resourceUsage(clock))
		}
		return true
	})
}

// CreatePod accepts the definition and try to start it.
func (node *Node) CreatePod(clock Time, pod *v1.Pod) error {
	log.L.Debugf("CreatePod(%v, %q) called", clock, pod.Name)

	key, err := buildKey(pod)
	if err != nil {
		return err
	}

	simSpec, err := parseSimSpec(pod)
	if err != nil {
		return err
	}

	simPod := simPod{pod: pod, startClock: clock, spec: simSpec}

	newTotalReq := sumResourceList(node.totalResourceRequest(clock), getResourceRequest(pod))
	capacity := node.config.Capacity
	if !greaterEqual(capacity, newTotalReq) || node.runningPodsNum(clock) >= capacity.Pods().Value() {
		simPod.status = simPodOverCapacity
	} else {
		simPod.status = simPodOk
	}

	node.pods.store(key, simPod)

	return nil
}

// func (node *Node) UpdatePod(clock Time, pod *v1.Pod) error {
// }

// GetPod returns a pod by name that was accepted on this node.
func (node *Node) GetPod(clock Time, namespace, name string) (*v1.Pod, error) {
	log.L.Debugf("GetPod(%v, %q, %q) called", clock, namespace, name)

	pod, err := node.getSimPod(namespace, name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, nil
	}

	return pod.pod, nil
}

// GetPods returns the list of all pods created on this node.
func (node *Node) GetPods(clock Time) ([]*v1.Pod, error) {
	log.L.Debugf("GetPods(%v) called", clock)
	return node.pods.listPods(), nil
}

// GetPodStatus returns the status of the pod by name
func (node *Node) GetPodStatus(clock Time, namespace, name string) (*v1.PodStatus, error) {
	log.L.Debugf("GetPodStatus(%v, %q, %q) called", clock, namespace, name)

	pod, err := node.getSimPod(namespace, name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, nil
	}

	status := pod.buildStatus(clock)

	return &status, nil
}

// DeletePod deletes the pod
func (node *Node) DeletePod(clock Time, pod *v1.Pod) error {
	log.L.Debugf("DeletePod(%v, %q) called", clock, pod.Name)

	key, err := buildKey(pod)
	if err != nil {
		return err
	}

	node.pods.remove(key)

	return nil
}

// TODO
// func (node *Node) NodeConditions(clock Time) []v1.NodeCondition {
// }

func (node *Node) totalResourceRequest(clock Time) v1.ResourceList {
	total := v1.ResourceList{}
	node.pods.foreach(func(_ string, pod simPod) bool {
		if pod.status == simPodOk && !pod.isTerminated(clock) {
			total = sumResourceList(total, getResourceRequest(pod.pod))
		}
		return true
	})
	return total
}

func (node *Node) runningPodsNum(clock Time) int64 {
	num := int64(0)
	node.pods.foreach(func(_ string, pod simPod) bool {
		if pod.status == simPodOk && !pod.isTerminated(clock) {
			num++
		}
		return true
	})
	return num
}

func (node *Node) getSimPod(namespace, name string) (*simPod, error) {
	key, err := buildKeyFromNames(namespace, name)
	if err != nil {
		return nil, err
	}

	pod, ok := node.pods.load(key)
	if !ok {
		return nil, nil
	}

	return &pod, nil
}

// buildKey builds a key for the provided pod.
func buildKey(pod *v1.Pod) (string, error) {
	if pod.ObjectMeta.Namespace == "" {
		return "", fmt.Errorf("pod namespace not found")
	}

	if pod.ObjectMeta.Name == "" {
		return "", fmt.Errorf("pod name not found")
	}

	return buildKeyFromNames(pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
}

func buildKeyFromNames(namespace string, name string) (string, error) {
	return fmt.Sprintf("%s-%s", namespace, name), nil
}
