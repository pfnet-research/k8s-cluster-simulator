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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
)

func TestBuildMetricsLogger(t *testing.T) {
	_, err := BuildMetricsLogger([]MetricsLoggerConfig{{
		Dest:      "",
		Formatter: "",
	}})
	assert.EqualError(t, err, "destination must not be empty")

	_, err = BuildMetricsLogger([]MetricsLoggerConfig{{
		Dest:      "foo",
		Formatter: "invalid",
	}})
	assert.EqualError(t, err, "formatter \"invalid\" is not supported")

	// TODO: Test correct cases
}

func TestBuildFormatter(t *testing.T) {
	actual0, _ := buildFormatter("JSON")
	expected0 := &metrics.JSONFormatter{}
	if actual0 != expected0 {
		t.Errorf("got: %+v\nwant: %+v", actual0, expected0)
	}

	actual1, _ := buildFormatter("humanReadable")
	expected1 := &metrics.HumanReadableFormatter{}
	if actual1 != expected1 {
		t.Errorf("got: %+v\nwant: %+v", actual1, expected1)
	}

	actual2, _ := buildFormatter("table")
	expected2 := &metrics.TableFormatter{}
	if actual2 != expected2 {
		t.Errorf("got: %+v\nwant: %+v", actual2, expected2)
	}

	_, err := buildFormatter("invalid")
	assert.EqualError(t, err, "formatter \"invalid\" is not supported")
}

func TestBuildNode(t *testing.T) {
	nowStr := time.Now().Format(time.RFC3339)
	nowParsed, _ := time.Parse(time.RFC3339, nowStr)

	metadata := metav1.ObjectMeta{
		Name: "node-0",
		Labels: map[string]string{
			"foo": "bar",
		},
		Annotations: map[string]string{},
	}

	spec := v1.NodeSpec{
		Unschedulable: false,
		Taints: []v1.Taint{
			{Key: "k", Value: "v", Effect: v1.TaintEffectNoSchedule},
		},
	}

	actual, _ := BuildNode(NodeConfig{
		Metadata: metadata,
		Spec:     spec,
		Status: NodeStatus{
			Allocatable: map[v1.ResourceName]string{
				"cpu":            "2",
				"memory":         "4Gi",
				"nvidia.com/gpu": "1",
			},
		},
	}, nowStr)

	allocatable := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	expected := v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metadata,
		Spec:       spec,
		Status: v1.NodeStatus{
			Capacity:    allocatable,
			Allocatable: allocatable,
			Conditions:  buildNodeCondition(metav1.NewTime(nowParsed)),
		},
	}

	if !reflect.DeepEqual(*actual, expected) {
		t.Errorf("got: %+v\nwant: %+v", *actual, expected)
	}
}

func TestBuildNodeConfig(t *testing.T) {
	now := metav1.NewTime(time.Now())

	actual := buildNodeCondition(now)
	expected := []v1.NodeCondition{
		{
			Type:               v1.NodeReady,
			Status:             v1.ConditionTrue,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletReady",
			Message:            "kubelet is posting ready status",
		},
		{
			Type:               v1.NodeOutOfDisk,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               v1.NodeMemoryPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               v1.NodeDiskPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               v1.NodePIDPressure,
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasSufficientPID",
			Message:            "kubelet has sufficient PID available",
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}
