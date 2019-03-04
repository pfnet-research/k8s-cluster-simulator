package config

import (
	"time"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// Config represents a user-specified simulator config.
type Config struct {
	LogLevel      string
	Tick          int
	StartClock    string
	MetricsFile   MetricsFileConfig
	MetricsStdout MetricsStdoutConfig
	// MetricsPort int
	// APIPort     int
	Cluster ClusterConfig
}

type MetricsFileConfig struct {
	Path      string
	Formatter string
}

type MetricsStdoutConfig struct {
	Formatter string
}

type ClusterConfig struct {
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

type TaintConfig struct { // made public for the deserialization by viper
	Key    string
	Value  string
	Effect string
}

// BuildMetricsFile builds a metrics.FileWriter with the given MetricsFileConfig.
// Returns error if the config is invalid, failed to parse, or failed to create a FileWriter.
func BuildMetricsFile(config MetricsFileConfig) (*metrics.FileWriter, error) {
	if config.Path == "" && config.Formatter == "" {
		return nil, nil
	}
	if config.Path == "" || config.Formatter == "" {
		return nil, strongerrors.InvalidArgument(errors.New("either metricsFile.Path or metricsFile.Formatter not given"))
	}

	formatter, err := buildFormatter(config.Formatter)
	if err != nil {
		return nil, err
	}

	return metrics.NewFileWriter(config.Path, formatter)
}

// BuildMetricsStdout builds a metrics.StdoutWriter with the given MetricsStdoutConfig.
// Returns error if parsing failed.
func BuildMetricsStdout(config MetricsStdoutConfig) (*metrics.StdoutWriter, error) {
	if config.Formatter == "" {
		return nil, nil
	}

	formatter, err := buildFormatter(config.Formatter)
	if err != nil {
		return nil, err
	}

	w := metrics.NewStdoutWriter(formatter)
	return &w, nil
}

func buildFormatter(config string) (metrics.Formatter, error) {
	switch config {
	case "JSON":
		return &metrics.JSONFormatter{}, nil
	case "humanReadable":
		return &metrics.HumanReadableFormatter{}, nil
	default:
		return nil, strongerrors.InvalidArgument(errors.Errorf("formatter %q is not supported", config))
	}
}

// BuildNode builds a *v1.Node with the provided node config.
// Returns error if the parsing fails.
func BuildNode(config NodeConfig, startClock string) (*v1.Node, error) {
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

	clock := time.Now()
	if startClock != "" {
		clock, err = time.Parse(time.RFC3339, startClock)
		if err != nil {
			return nil, err
		}
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
			Conditions:  buildNodeCondition(metav1.NewTime(clock)),
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
