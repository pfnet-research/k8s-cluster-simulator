package sim

import (
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/log"
)

// Node represents a simulated computing node.
type Node struct {
	config        NodeConfig
	pods          podMap
	resourceUsage v1.ResourceList
}

// NodeConfig represents configuration of this node.
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

// UpdateState updates the state of this node.
func (node *Node) UpdateState(clock Time) {
	log.L.Debugf("Node %q: UpdateState(%v) called", node.config.Name, clock)

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
	log.L.Debugf("Node %q: CreatePod(%v, %q) called", node.config.Name, clock, pod.Name)

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

// GetPod returns a pod by name that was accepted on this node.
// The returned pod may have failed to be scheduled.
func (node *Node) GetPod(clock Time, namespace, name string) (*v1.Pod, error) {
	log.L.Debugf("Node %q: GetPod(%v, %q, %q) called", node.config.Name, clock, namespace, name)

	pod, err := node.getSimPod(namespace, name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, nil
	}

	return pod.pod, nil
}

// GetPodList returns the list of all pods that were accepted on this node.
// Each of the returned pods may have failed to be scheduled.
func (node *Node) GetPodList(clock Time) []*v1.Pod {
	log.L.Debugf("Node %q: GetPodList(%v) called", node.config.Name, clock)
	return node.pods.listPods()
}

// GetPodStatus returns the status of the pod by name.
func (node *Node) GetPodStatus(clock Time, namespace, name string) (*v1.PodStatus, error) {
	log.L.Debugf("Node %q: GetPodStatus(%v, %q, %q) called", node.config.Name, clock, namespace, name)

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
