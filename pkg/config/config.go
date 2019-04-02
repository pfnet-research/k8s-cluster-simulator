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

package config

import (
	"time"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// Config represents a user-specified simulator config.
type Config struct {
	LogLevel      string
	Tick          int
	StartClock    string
	MetricsTick   int
	MetricsLogger []MetricsLoggerConfig
	Cluster       []NodeConfig
}

// Made public to be parsed from YAML.

type MetricsLoggerConfig struct {
	// Dest is an output device or file path in which the metrics is written.
	Dest string
	// Formatter is a type of metrics format.
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

// BuildMetricsLogger builds metrics.FileWriter with the given MetricsLoggerConfig.
// Returns error if the config is invalid or failed to create a FileWriter.
func BuildMetricsLogger(conf []MetricsLoggerConfig) ([]*metrics.FileWriter, error) {
	writers := make([]*metrics.FileWriter, 0, len(conf))

	for _, conf := range conf {
		if conf.Dest == "" {
			return nil, strongerrors.InvalidArgument(errors.New("destination must not be empty"))
		}

		formatter, err := buildFormatter(conf.Formatter)
		if err != nil {
			return nil, err
		}

		writer, err := metrics.NewFileWriter(conf.Dest, formatter)
		if err != nil {
			return nil, err
		}

		writers = append(writers, writer)
	}

	return writers, nil
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

// BuildNode builds a *v1.Node with the given NodeConfig.
// Returns error if failed to parse.
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
			Message:            "kubelet is posting ready status",
		},
		{
			Type:               v1.NodeOutOfDisk,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               v1.NodeMemoryPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               v1.NodeDiskPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               v1.NodePIDPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  clock,
			LastTransitionTime: clock,
			Reason:             "KubeletHasSufficientPID",
			Message:            "kubelet has sufficient PID available",
		},
	}
}
