package config

import (
	"time"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/k8s-cluster-simulator/kubesim/metrics"
	"github.com/ordovicia/k8s-cluster-simulator/kubesim/util"
)

// Config represents a user-specified simulator config.
type Config struct {
	LogLevel      string
	Tick          int
	StartClock    string
	MetricsTick   int
	MetricsFile   []MetricsFileConfig
	MetricsStdout MetricsStdoutConfig
	Cluster       []NodeConfig
}

type MetricsFileConfig struct {
	Path      string
	Formatter string
}

type MetricsStdoutConfig struct {
	Formatter string
}

type NodeConfig struct {
	Metadata metav1.ObjectMeta
	Spec     v1.NodeSpec
	Status   NodeStatus
}

type NodeStatus struct {
	Allocatable map[v1.ResourceName]string
}

// BuildMetricsFile builds metrics.FileWriter with the given MetricsFileConfig.
// Returns error if the config is invalid, failed to parse, or failed to create a FileWriter.
func BuildMetricsFile(conf []MetricsFileConfig) ([]*metrics.FileWriter, error) {
	writers := make([]*metrics.FileWriter, 0, len(conf))

	for _, conf := range conf {
		if conf.Path == "" && conf.Formatter == "" {
			return nil, nil
		}
		if conf.Path == "" || conf.Formatter == "" {
			return nil, strongerrors.InvalidArgument(errors.New("either metricsFile.Path or metricsFile.Formatter not given"))
		}

		formatter, err := buildFormatter(conf.Formatter)
		if err != nil {
			return nil, err
		}

		writer, err := metrics.NewFileWriter(conf.Path, formatter)
		if err != nil {
			return nil, err
		}

		writers = append(writers, writer)
	}

	return writers, nil
}

// BuildMetricsStdout builds a metrics.StdoutWriter with the given MetricsStdoutConfig.
// Returns error if parsing failed.
func BuildMetricsStdout(conf MetricsStdoutConfig) (*metrics.StdoutWriter, error) {
	if conf.Formatter == "" {
		return nil, nil
	}

	formatter, err := buildFormatter(conf.Formatter)
	if err != nil {
		return nil, err
	}

	w := metrics.NewStdoutWriter(formatter)
	return &w, nil
}

func buildFormatter(conf string) (metrics.Formatter, error) {
	switch conf {
	case "JSON":
		return &metrics.JSONFormatter{}, nil
	case "humanReadable":
		return &metrics.HumanReadableFormatter{}, nil
	case "table":
		return &metrics.TableFormatter{}, nil
	default:
		return nil, strongerrors.InvalidArgument(errors.Errorf("formatter %q is not supported", conf))
	}
}

// BuildNode builds a *v1.Node with the given node config.
// Returns error if the parsing fails.
func BuildNode(conf NodeConfig, startClock string) (*v1.Node, error) {
	allocatable, err := util.BuildResourceList(conf.Status.Allocatable)
	if err != nil {
		return nil, err
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
		ObjectMeta: conf.Metadata,
		Spec:       conf.Spec,
		Status: v1.NodeStatus{
			Capacity:    allocatable,
			Allocatable: allocatable,
			Conditions:  buildNodeCondition(metav1.NewTime(clock)),
		},
	}

	return &node, nil
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
