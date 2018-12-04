package kubesim

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/ordovicia/kubernetes-simulator/sim"
)

// Config represents a simulator config by user
type Config struct {
	Cluster     ClusterConfig
	Tick        int
	APIPort     int
	MetricsPort int
	LogLevel    string
}

type ClusterConfig struct {
	Nodes []NodeConfig
}

type NodeConfig struct {
	Name     string
	Capacity map[v1.ResourceName]string
	Labels   map[string]string // TODO: force constraints
	Taints   []TaintConfig
}

type TaintConfig struct {
	Key    string // TODO: force constraints
	Value  string
	Effect string
}

func buildNodeConfig(config NodeConfig) (*sim.NodeConfig, error) {
	capacity, err := buildCapacity(config.Capacity)
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

	return &sim.NodeConfig{
		Name:     config.Name,
		Capacity: capacity,
		Labels:   config.Labels,
		Taints:   taints,
	}, nil
}

func buildCapacity(config map[v1.ResourceName]string) (v1.ResourceList, error) {
	resourceList := v1.ResourceList{}

	for key, value := range config {
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return nil, strongerrors.InvalidArgument(errors.Errorf("invalid %s value %q", key, value))
		}
		resourceList[key] = quantity
	}

	return resourceList, nil
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
