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

package node

import (
	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// Node represents a simulated computing node.
type Node struct {
	v1   *v1.Node
	pods map[string]*pod.Pod
}

// Metrics is a metrics of a Node at one point of time.
type Metrics struct {
	Allocatable          v1.ResourceList
	RunningPodsNum       int64
	TerminatingPodsNum   int64
	FailedPodsNum        int64
	TotalResourceRequest v1.ResourceList
	TotalResourceUsage   v1.ResourceList
}

// NewNode creates a new Node with the given v1.Node.
func NewNode(node *v1.Node) Node {
	return Node{
		v1:   node,
		pods: map[string]*pod.Pod{},
	}
}

// ToV1 returns *v1.Node representation of this Node.
func (node *Node) ToV1() *v1.Node {
	return node.v1
}

// ToNodeInfo creates *nodeinfo.NodeInfo object from this Node.
func (node *Node) ToNodeInfo(clock clock.Clock) (*nodeinfo.NodeInfo, error) {
	pods := node.runningAndTerminatingPodsV1WithStatus(clock)
	nodeInfo := nodeinfo.NewNodeInfo(pods...)
	err := nodeInfo.SetNode(node.ToV1())
	if err != nil {
		return nil, err
	}
	return nodeInfo, nil
}

// Metrics returns the Metrics of this Node at the given clock.
func (node *Node) Metrics(clock clock.Clock) Metrics {
	return Metrics{
		Allocatable:          node.ToV1().Status.Allocatable,
		RunningPodsNum:       node.runningPodsNum(clock),
		TerminatingPodsNum:   node.terminatingPodsNum(clock),
		FailedPodsNum:        node.bindingFailedPodsNum(),
		TotalResourceRequest: node.totalResourceRequest(clock),
		TotalResourceUsage:   node.totalResourceUsage(clock),
	}
}

// BindPod accepts the given pod and try to start it.
// The pod will fail to be started if there is not sufficient resources.
// Returns the bound pod in pod.Pod representation, or error if the pod has invalid name or failed
// to create a simulated pod.
func (node *Node) BindPod(clock clock.Clock, v1Pod *v1.Pod) (*pod.Pod, error) {
	key, err := util.PodKey(v1Pod)
	if err != nil {
		return nil, err
	}

	log.L.Tracef("Node %s: Pod %s bound", node.ToV1().Name, key)

	// Check node capacity
	newTotalReq := util.ResourceListSum(node.totalResourceRequest(clock), util.PodTotalResourceRequests(v1Pod))
	allocatable := node.ToV1().Status.Allocatable
	var podStatus pod.Status

	if !util.ResourceListGE(allocatable, newTotalReq) || node.runningPodsNum(clock) >= allocatable.Pods().Value() {
		podStatus = pod.OverCapacity
	} else {
		podStatus = pod.Ok
	}

	// Create simulated pod
	simPod, err := pod.NewPod(v1Pod, clock, podStatus, node.ToV1().Name)
	if err != nil {
		return nil, err
	}
	v1Pod.Status = simPod.BuildStatus(clock)
	node.pods[key] = simPod

	return simPod, nil
}

// DeletePod start deleting the given pod from this Node.
// Returns true if the pod is found in this Node, or false otherwise.
func (node *Node) DeletePod(clock clock.Clock, podNamespace, podName string) bool {
	key := util.PodKeyFromNames(podNamespace, podName)
	pod, ok := node.pods[key]
	if ok {
		pod.Delete(clock)
	}

	return ok
}

// Pod returns the *pod.Pod by name that was accepted on this node.
// The returned pod may have failed to be started.
// Returns nil if the pod is not found.
func (node *Node) Pod(namespace, name string) *pod.Pod {
	key := util.PodKeyFromNames(namespace, name)
	pod, ok := node.pods[key]
	if !ok {
		return nil
	}

	return pod
}

// PodList returns a list of all pods that were accepted on this Node and not terminated nor
// deleted.
// Each of the returned pods may have failed to be started.
func (node *Node) PodList() []*pod.Pod {
	podList := make([]*pod.Pod, 0, len(node.pods))
	for _, pod := range node.pods {
		podList = append(podList, pod)
	}

	return podList
}

// PodsNum returns the number of all running or terminating pods on this Node at the given clock.
func (node *Node) PodsNum(clock clock.Clock) int64 {
	return node.runningPodsNum(clock) + node.terminatingPodsNum(clock)
}

// GCTerminatedPods deletes terminated or deleted pods at the given clock from this Node.
func (node *Node) GCTerminatedPods(clock clock.Clock) {
	for name, pod := range node.pods {
		if pod.IsTerminated(clock) || pod.IsDeleted(clock) {
			delete(node.pods, name)
		}
	}
}

// runningAndTerminatingPodsV1WithStatus returns all running or terminating pods on this Node in
// *v1.Pod representation at the given clock, with their status updated.
func (node *Node) runningAndTerminatingPodsV1WithStatus(clock clock.Clock) []*v1.Pod {
	podList := []*v1.Pod{}
	for _, pod := range node.pods {
		if pod.IsRunning(clock) || pod.IsTerminating(clock) {
			podV1 := pod.ToV1()
			podV1.Status = pod.BuildStatus(clock)
			podList = append(podList, podV1)
		}
	}

	return podList
}

// totalResourceRequest calculates the total resource request (not usage) of all running or
// terminating pods on this Node at the given clock.
func (node *Node) totalResourceRequest(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	for _, pod := range node.pods {
		if pod.IsRunning(clock) || pod.IsTerminating(clock) {
			total = util.ResourceListSum(total, pod.TotalResourceRequests())
		}
	}

	return total
}

// runningPodsNum returns the number of all running pods no this Node at the given clock.
func (node *Node) runningPodsNum(clock clock.Clock) int64 {
	num := int64(0)
	for _, pod := range node.pods {
		if pod.IsRunning(clock) {
			num++
		}
	}

	return num
}

// terminatingPodsNum returns the number of all terminating pods no this Node at the given clock.
func (node *Node) terminatingPodsNum(clock clock.Clock) int64 {
	num := int64(0)
	for _, pod := range node.pods {
		if pod.IsTerminating(clock) {
			num++
		}
	}

	return num
}

// bindingFailedPodsNum returns the number of pods that failed to be started on this Node.
func (node *Node) bindingFailedPodsNum() int64 {
	num := int64(0)
	for _, pod := range node.pods {
		if pod.HasFailedToStart() {
			num++
		}
	}

	return num
}

// totalResourceUsage calculates the total resource usage (not request) of all running or
// terminating pods at the given clock.
func (node *Node) totalResourceUsage(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	for _, pod := range node.pods {
		if pod.IsRunning(clock) || pod.IsTerminating(clock) {
			total = util.ResourceListSum(total, pod.ResourceUsage(clock))
		}
	}

	return total
}
