package node

import (
	"errors"
	"fmt"

	"github.com/cpuguy83/strongerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// Node represents a simulated computing node.
type Node struct {
	v1   *v1.Node
	pods pod.Map
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
	pods := node.runningPodsWithStatus(clock)
	nodeInfo := nodeinfo.NewNodeInfo(pods...)
	nodeInfo.SetNode(node.ToV1())
	return nodeInfo
}

// CreatePod accepts the definition of a pod and try to start it. The pod will fail to be scheduled
// if there is not sufficient resources.
func (node *Node) CreatePod(clock clock.Clock, v1Pod *v1.Pod) error {
	log.L.Tracef("Node %s: Pod %s bound", node.ToV1().Name, v1Pod.Name)

	key, err := buildKey(v1Pod)
	if err != nil {
		return err
	}

	newTotalReq := resourceListSum(node.totalResourceRequest(clock), extractResourceRequest(v1Pod))
	capacity := node.ToV1().Status.Capacity
	var podStatus pod.Status
	if !resourceListGE(capacity, newTotalReq) || node.runningPodsNum(clock) >= capacity.Pods().Value() {
		podStatus = pod.OverCapacity
	} else {
		podStatus = pod.Ok
	}

	simPod, err := pod.NewPod(v1Pod, clock, podStatus)
	if err != nil {
		return err
	}

	node.pods.Store(key, *simPod)
	return nil
}

// Pod returns the pod by name that was accepted on this node. The returned pod may have failed to
// be scheduled. Returns error if the pod is not found.
func (node *Node) Pod(namespace, name string) (*v1.Pod, error) {
	pod := node.simPod(namespace, name)
	if pod == nil {
		return nil, strongerrors.NotFound(fmt.Errorf("pod %q not found", buildKeyFromNames(namespace, name)))
	}

	return pod.ToV1(), nil
}

// PodList returns the list of all pods that were accepted on this node. Each of the returned pods
// may have failed to be scheduled.
func (node *Node) PodList() []*v1.Pod {
	return node.pods.ListPods()
}

// runningPodsWithStatus returns all running pods at the time clock, with their status updated.
func (node *Node) runningPodsWithStatus(clock clock.Clock) []*v1.Pod {
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
			total = resourceListSum(total, extractResourceRequest(pod.ToV1()))
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

// simPod returns a *pod.Pod by name that was accepted on this node.
// The returned pod may have failed to be scheduled.
// Returns nil if the pod is not found.
func (node *Node) simPod(namespace, name string) *pod.Pod {
	key := buildKeyFromNames(namespace, name)
	pod, ok := node.pods.Load(key)
	if !ok {
		return nil
	}

	return pod
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
