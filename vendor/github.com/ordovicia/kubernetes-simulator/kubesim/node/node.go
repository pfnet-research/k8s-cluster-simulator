package node

import (
	"errors"
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/cpuguy83/strongerrors"
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

// CreatePod accepts the definition of a pod and try to start it.
// The pod will fail to be scheduled if there is not sufficient resources.
func (node *Node) CreatePod(clock clock.Clock, v1Pod *v1.Pod) error {
	log.L.Debugf("Node %q: CreatePod(%v, %q) called", node.v1.Name, clock, v1Pod.Name)

	key, err := buildKey(v1Pod)
	if err != nil {
		return err
	}

	newTotalReq := resourceListSum(node.totalResourceRequest(clock), getResourceReq(v1Pod))
	cap := node.v1.Status.Capacity
	var podStatus pod.Status
	if !resourceListGE(cap, newTotalReq) || node.runningPodsNum(clock) >= cap.Pods().Value() {
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

// GetPod returns a pod by name that was accepted on this node.
// The returned pod may have failed to be scheduled.
func (node *Node) GetPod(clock clock.Clock, namespace, name string) (*v1.Pod, error) {
	log.L.Debugf("Node %q: GetPod(%v, %q, %q) called", node.v1.Name, clock, namespace, name)

	pod := node.getSimPod(namespace, name)
	if pod == nil {
		return nil, strongerrors.NotFound(fmt.Errorf("pod %q not found", buildKeyFromNames(namespace, name)))
	}

	return pod.ToV1(), nil
}

// GetPodList returns the list of all pods that were accepted on this node.
// Each of the returned pods may have failed to be scheduled.
func (node *Node) GetPodList(clock clock.Clock) []*v1.Pod {
	log.L.Debugf("Node %q: GetPodList(%v) called", node.v1.Name, clock)
	return node.pods.ListPods()
}

// GetPodStatus returns the status of the pod by name.
func (node *Node) GetPodStatus(clock clock.Clock, namespace, name string) (*v1.PodStatus, error) {
	log.L.Debugf("Node %q: GetPodStatus(%v, %q, %q) called", node.v1.Name, clock, namespace, name)

	pod := node.getSimPod(namespace, name)
	if pod == nil {
		return nil, strongerrors.NotFound(fmt.Errorf("pod %q not found", buildKeyFromNames(namespace, name)))
	}

	status := pod.BuildStatus(clock)
	return &status, nil
}

// totalResourceRequest calculates the total resource request (not usage) of running pods at the
// time clock.
func (node *Node) totalResourceRequest(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsRunning(clock) {
			total = resourceListSum(total, getResourceReq(pod.ToV1()))
		}
		return true
	})
	return total
}

// runningPodsNum returns the number of running pods at the time clock.
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

// getSimPod returns a *pod.Pod by name that was accepted on this node.
// The returned pod may have failed to be scheduled.
func (node *Node) getSimPod(namespace, name string) *pod.Pod {
	key := buildKeyFromNames(namespace, name)
	pod, ok := node.pods.Load(key)
	if !ok {
		return nil
	}

	return pod
}

// buildKey builds a key for the provided pod.
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
	return fmt.Sprintf("%s-%s", namespace, name)
}
