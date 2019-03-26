package config

import (
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ordovicia/k8s-cluster-simulator/pkg/metrics"
)

func TestBuildMetricsFile(t *testing.T) {
	actual, err := BuildMetricsFile([]MetricsFileConfig{MetricsFileConfig{
		Path:      "",
		Formatter: "",
	}})
	if actual != nil || err != nil {
		t.Errorf("got: (%+v, %+v)\nwant: (nil, nil)", actual, err)
	}

	actual, err = BuildMetricsFile([]MetricsFileConfig{MetricsFileConfig{
		Path:      "",
		Formatter: "foo",
	}})
	if err == nil {
		t.Errorf("got: nil\nwant: error")
	}

	actual, err = BuildMetricsFile([]MetricsFileConfig{MetricsFileConfig{
		Path:      "foo",
		Formatter: "",
	}})
	if err == nil {
		t.Errorf("got: nil\nwant: error")
	}

	actual, err = BuildMetricsFile([]MetricsFileConfig{
		MetricsFileConfig{
			Path:      "",
			Formatter: "",
		},
		MetricsFileConfig{
			Path:      "",
			Formatter: "",
		},
	})
	if err != nil {
		t.Errorf("got: nil\nwant: error")
	}
}

func TestBuildMetricsStdout(t *testing.T) {
	actual, err := BuildMetricsStdout(MetricsStdoutConfig{
		Formatter: "",
	})
	if actual != nil || err != nil {
		t.Errorf("got: (%+v, %+v)\nwant: (nil, nil)", actual, err)
	}

	_, err = BuildMetricsStdout(MetricsStdoutConfig{
		Formatter: "foo",
	})
	if err == nil {
		t.Errorf("got: nil\nwant: error")
	}

	_, err = BuildMetricsStdout(MetricsStdoutConfig{
		Formatter: "JSON",
	})
	if err != nil {
		t.Errorf("got: %+v\nwant: metrics.StdoutWriter", err)
	}
}

func TestBuildFormatter(t *testing.T) {
	actual, _ := buildFormatter("JSON")
	expected := &metrics.JSONFormatter{}
	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	_, err := buildFormatter("Invalid")
	if err == nil {
		t.Errorf("got: nil\nwant: error")
	}
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
			v1.Taint{Key: "k", Value: "v", Effect: v1.TaintEffectNoSchedule},
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
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
