package config

import (
	"testing"

	"k8s.io/api/core/v1"
)

/*
func TestBuildNode(t *testing.T) {
	now := time.Now()

	capacity := map[v1.ResourceName]string{
		"cpu":            "2",
		"memory":         "4Gi",
		"nvidia.com/gpu": "1",
	}

	actual, _ := BuildNode(NodeConfig{
		Namespace:   "default",
		Name:        "node-00",
		Capacity:    capacity,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		Taints:      []TaintConfig{},
	}, now.String())

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
			Name:        "node-00",
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
			Conditions:  buildNodeCondition(metav1.NewTime(now)),
		},
	}

	if reflect.DeepEqual(*actual, expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
*/

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
