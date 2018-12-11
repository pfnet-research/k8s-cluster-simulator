package node

import (
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/pod"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// Node represents a simulated computing node.
type Node struct {
	config        Config
	pods          pod.Map
	resourceUsage v1.ResourceList
}

// Config represents configuration of this node.
type Config struct {
	Name     string
	Capacity v1.ResourceList
	Labels   map[string]string // TODO
	Taints   []v1.Taint        // TODO
}

// NewNode creates a new node with the config.
func NewNode(config Config) Node {
	return Node{
		config:        config,
		pods:          pod.Map{},
		resourceUsage: v1.ResourceList{},
	}
}

// CreatePod accepts the definition of a pod and try to start it.
// The pod will fail to be scheduled if there is not sufficient resources.
func (node *Node) CreatePod(clock clock.Clock, podDef *v1.Pod) error {
	log.L.Debugf("Node %q: CreatePod(%v, %q) called", node.config.Name, clock, podDef.Name)

	key, err := buildKey(podDef)
	if err != nil {
		return err
	}

	newTotalReq := sumResourceList(node.totalResourceRequest(clock), getResourceRequest(podDef))
	cap := node.config.Capacity
	var podStatus pod.Status
	if !greaterEqual(cap, newTotalReq) || node.runningPodsNum(clock) >= cap.Pods().Value() {
		podStatus = pod.OverCapacity
	} else {
		podStatus = pod.Ok
	}

	simPod, err := pod.NewPod(podDef, clock, podStatus)
	if err != nil {
		return err
	}

	node.pods.Store(key, *simPod)
	return nil
}

// GetPod returns a pod by name that was accepted on this node.
// The returned pod may have failed to be scheduled.
func (node *Node) GetPod(clock clock.Clock, namespace, name string) (*v1.Pod, error) {
	log.L.Debugf("Node %q: GetPod(%v, %q, %q) called", node.config.Name, clock, namespace, name)

	pod, err := node.getSimPod(namespace, name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, nil
	}

	return pod.ToV1(), nil
}

// GetPodList returns the list of all pods that were accepted on this node.
// Each of the returned pods may have failed to be scheduled.
func (node *Node) GetPodList(clock clock.Clock) []*v1.Pod {
	log.L.Debugf("Node %q: GetPodList(%v) called", node.config.Name, clock)
	return node.pods.ListPods()
}

// GetPodStatus returns the status of the pod by name.
func (node *Node) GetPodStatus(clock clock.Clock, namespace, name string) (*v1.PodStatus, error) {
	log.L.Debugf("Node %q: GetPodStatus(%v, %q, %q) called", node.config.Name, clock, namespace, name)

	pod, err := node.getSimPod(namespace, name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, nil
	}

	status := pod.BuildStatus(clock)

	return &status, nil
}

// TODO
// func (node *Node) NodeConditions(clock clock.Clock) []v1.NodeCondition {
// }

// updateState updates the state of this node.
// TODO
func (node *Node) updateState(clock clock.Clock) {
	log.L.Debugf("Node %q: UpdateState(%v) called", node.config.Name, clock)

	node.resourceUsage = v1.ResourceList{}
	node.pods.Range(func(key string, pod pod.Pod) bool {
		if pod.IsTerminated(clock) {
			// TODO
		} else {
			node.resourceUsage = sumResourceList(node.resourceUsage, pod.ResourceUsage(clock))
		}
		return true
	})
}

func (node *Node) totalResourceRequest(clock clock.Clock) v1.ResourceList {
	total := v1.ResourceList{}
	node.pods.Range(func(_ string, pod pod.Pod) bool {
		if pod.IsRunning(clock) {
			total = sumResourceList(total, getResourceRequest(pod.ToV1()))
		}
		return true
	})
	return total
}

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

func (node *Node) getSimPod(namespace, name string) (*pod.Pod, error) {
	key, err := buildKeyFromNames(namespace, name)
	if err != nil {
		return nil, err
	}

	pod, ok := node.pods.Load(key)
	if !ok {
		return nil, nil
	}

	return pod, nil
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
