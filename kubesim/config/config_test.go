package config

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildNode(t *testing.T) {
	nowStr := time.Now().Format(time.RFC3339)
	nowParsed, _ := time.Parse(time.RFC3339, nowStr)

	actual, _ := BuildNode(NodeConfig{
		Namespace: "default",
		Name:      "node0",
		Capacity: map[v1.ResourceName]string{
			"cpu":            "2",
			"memory":         "4Gi",
			"nvidia.com/gpu": "1",
		},
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		Taints:      []TaintConfig{},
	}, nowStr)

	cap := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	expected := v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "node0",
			Namespace:   "default",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: v1.NodeSpec{
			Unschedulable: false,
			Taints:        []v1.Taint{},
		},
		Status: v1.NodeStatus{
			Capacity:    cap,
			Allocatable: cap,
			Conditions:  buildNodeCondition(metav1.NewTime(nowParsed)),
		},
	}

	if !reflect.DeepEqual(*actual, expected) {
		t.Errorf("got: %+v\nwant: %+v", *actual, expected)
	}
}

func TestBuildTaint(t *testing.T) {
	actual, _ := buildTaint(TaintConfig{
		Key:    "kubernetes",
		Value:  "simulator",
		Effect: "NoSchedule",
	})

	expected := v1.Taint{
		Key:    "kubernetes",
		Value:  "simulator",
		Effect: v1.TaintEffectNoSchedule,
	}

	if *actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	actual, err := buildTaint(TaintConfig{
		Key:    "kubernetes",
		Value:  "simulator",
		Effect: "Invalid",
	})

	if err == nil {
		t.Errorf("got: %v\nwant: error", actual)
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
			Message:            "kubelet is ready.",
		},
		{
			Type:               "OutOfDisk",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               "MemoryPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               "DiskPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               "NetworkUnavailable",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  now,
			LastTransitionTime: now,
			Reason:             "RouteCreated",
			Message:            "RouteController created a route",
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
