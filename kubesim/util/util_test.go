package util_test

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

func TestBuildResourceList(t *testing.T) {
	rsrc := map[v1.ResourceName]string{
		"cpu":            "1",
		"memory":         "2Gi",
		"nvidia.com/gpu": "1",
	}

	expected := v1.ResourceList{
		"cpu":            resource.MustParse("1"),
		"memory":         resource.MustParse("2Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	actual, _ := util.BuildResourceList(rsrc)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("got: %#v\nwant: %#v", actual, expected)
	}

	rsrc = map[v1.ResourceName]string{
		"cpu":    "1",
		"memory": "2Gi",
		"foo":    "bar",
	}

	actual, err := util.BuildResourceList(rsrc)
	if err == nil {
		t.Errorf("got: %v\nwant: error", actual)
	}
}
