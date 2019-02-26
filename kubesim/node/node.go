package node

import (
	"errors"
	"fmt"

	"github.com/cpuguy83/strongerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// Node represents a simulated computing node.
type Node struct {
	v1   *v1.Node
	pods pod.Map
}

// Metrics is a metrics of a node at a time instance.
type Metrics struct {
	Capacity             v1.ResourceList
	RunningPodsNum       int64
	FailedPodsNum        int64
	TotalResourceRequest v1.ResourceList
	TotalResourceUsage   v1.ResourceList
}

// NewNode creates a new node with the v1.Node definition.
func NewNode(node *v1.Node) Node {
	return Node{
		v1:   node,
		pods: pod.Map{},
	}
}

// ToV1 returns v1.Node representation of this node.
func (node *Node) ToV1() *v1.Node {
	return node.v1
}

// ToNodeInfo creates nodeinfo.NodeInfo object from this node.
func (node *Node) ToNodeInfo(clock clock.Clock) *nodeinfo.NodeInfo {
	pods := node.runningV1PodsWithStatus(clock)
	nodeInfo := nodeinfo.NewNodeInfo(pods...)
	nodeInfo.SetNode(node.ToV1())
	return nodeInfo
}

// Metrics returns the Metrics at the time clock.
func (node *Node) Metrics(clock clock.Clock) Metrics {
	return Metrics{
		Capacity:             node.ToV1().Status.Capacity,
		RunningPodsNum:       node.runningPodsNum(clock),
		FailedPodsNum:        node.bindingFailedPodsNum(),
		TotalResourceRequest: node.totalResourceRequest(clock),
		TotalResourceUsage:   node.totalResourceUsage(clock),
	}
}

// CreatePod accepts the definition of a pod and try to start it. The pod will fail to be bound if
// there is not sufficient resources.
func (node *Node) CreatePod(clock clock.Clock, v1Pod *v1.Pod) error {
	log.L.Tracef("Node %s: Pod %s bound", node.ToV1().Name, v1Pod.Name)

	key, err := buildKey(v1Pod)
	if err != nil {
		return err
	}

	newTotalReq := util.ResourceListSum(node.totalResourceRequest(clock), util.PodTotalResourceRequests(v1Pod))
	capacity := node.ToV1().Status.Capacity
	var podStatus pod.Status
	if !util.ResourceListGE(capacity, newTotalReq) || node.runningPodsNum(clock) >= capacity.Pods().Value() {
		podStatus = pod.OverCapacity
	} else {
		podStatus = pod.Ok
	}

	simPod, err := pod.NewPod(v1Pod, clock, podStatus, node.ToV1().Name)
	if err != nil {
		return err
	}

	node.pods.Store(key, *simPod)
	return nil
}

// Pod returns the *pod.Pod by name that was accepted on this node. The returned pod may have
// failed to be bound. Returns nil if the pod is not found.
func (node *Node) Pod(namespace, name string) *pod.Pod {
	key := buildKeyFromNames(namespace, name)
	pod, ok := node.pods.Load(key)
	if !ok {
		return nil
	}

	return pod
}

// PodList returns the list of all pods that were accepted on this node. Each of the returned pods
// may have failed to be bound.
func (node *Node) PodList() []pod.Pod {
	return node.pods.ListPods()
}

// runningV1PodsWithStatus returns all running pods in *v1.Pod representation at the time clock,
// with their status updated.
func (node *Node) runningV1PodsWithStatus(clock clock.Clock) []*v1.Pod {
	pods := []*v1.Pod{}
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		podV1 := pod.ToV1()
		podV1.Status = pod.BuildStatus(clock)
		if pod.IsRunning(clock) {
			pods = append(pods, podV1)
		}
		return true
	})

	return pods
}

// totalResourceRequest calculates the total resource request (not usage) of all running pods at the
// time clock.
func (node *Node) totalResourceRequest(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsRunning(clock) {
			total = util.ResourceListSum(total, pod.TotalResourceRequests())
		}
		return true
	})

	return total
}

// runningPodsNum returns the number of all running pods at the time clock.
func (node *Node) runningPodsNum(clock clock.Clock) int64 {
	num := int64(0)
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsRunning(clock) {
			num++
		}
		return true
	})
	return num
}

// bindingFailedPodsNum returns the number of pods that failed to be bound to this node.
func (node *Node) bindingFailedPodsNum() int64 {
	num := int64(0)
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsBindingFailed() {
			num++
		}
		return true
	})
	return num
}

// totalResourceUsage calculates the total resource usage of all running pods at the time clock.
func (node *Node) totalResourceUsage(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsRunning(clock) {
			total = util.ResourceListSum(total, pod.ResourceUsage(clock))
		}
		return true
	})

	return total
}

// buildKey builds a key for the provided pod.
// Returns error if the pod does not have valid (= non-empty) namespace and name.
func buildKey(pod *v1.Pod) (string, error) {
	if pod.ObjectMeta.Namespace == "" {
		return "", strongerrors.InvalidArgument(errors.New("Empty pod namespace"))
	}

	if pod.ObjectMeta.Name == "" {
		return "", strongerrors.InvalidArgument(errors.New("Empty pod name"))
	}

	return buildKeyFromNames(pod.ObjectMeta.Namespace, pod.ObjectMeta.Name), nil
}

// buildKeyFromNames builds a key from the namespace and pod name.
func buildKeyFromNames(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
