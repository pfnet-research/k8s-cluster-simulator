package kubesim

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// Config represents a simulator config by user.
type Config struct {
	LogLevel string
	Tick     int
	// APIPort     int
	// MetricsPort int
	Cluster ClusterConfig
}

type ClusterConfig struct { // publicized for deserialization by viper
	Nodes []NodeConfig
}

type NodeConfig struct {
	Namespace   string
	Name        string
	Capacity    map[v1.ResourceName]string
	Labels      map[string]string
	Annotations map[string]string
	Taints      []TaintConfig
}

type TaintConfig struct {
	Key    string
	Value  string
	Effect string
}

// buildNode builds a *v1.Node with the provided node config.
func buildNode(config NodeConfig) (*v1.Node, error) {
	capacity, err := util.BuildResourceList(config.Capacity)
	if err != nil {
		return nil, err
	}

	taints := []v1.Taint{}
	for _, taintConfig := range config.Taints {
		taint, err := buildTaint(taintConfig)
		if err != nil {
			return nil, err
		}
		taints = append(taints, *taint)
	}

	node := v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.Name,
			Namespace:   config.Namespace,
			Labels:      config.Labels,
			Annotations: config.Annotations,
		},
		Spec: v1.NodeSpec{
			Unschedulable: false,
			Taints:        taints,
		},
		Status: v1.NodeStatus{
			Capacity:    capacity,
			Allocatable: capacity,
			Conditions:  buildNodeCondition(metav1.Now()),
		},
	}

	return &node, nil
}

func buildTaint(config TaintConfig) (*v1.Taint, error) {
	var effect v1.TaintEffect
	switch config.Effect {
	case "NoSchedule":
		effect = v1.TaintEffectNoSchedule
	case "NoExecute":
		effect = v1.TaintEffectNoExecute
	case "PreferNoSchedule":
		effect = v1.TaintEffectPreferNoSchedule
	default:
		return nil, strongerrors.InvalidArgument(errors.Errorf("taint effect %q is not supported", config.Effect))
	}

	return &v1.Taint{
		Key:    config.Key,
		Value:  config.Value,
		Effect: effect,
	}, nil
}

func buildNodeCondition(clock metav1.Time) []v1.NodeCondition {
	return []v1.NodeCondition{
		{
			Type:               v1.NodeReady,
			Status:             v1.ConditionTrue,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletReady",
			Message:            "kubelet is ready.",
		},
		{
			Type:               "OutOfDisk",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               "MemoryPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               "DiskPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               "NetworkUnavailable",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "RouteCreated",
			Message:            "RouteController created a route",
		},
	}
}
