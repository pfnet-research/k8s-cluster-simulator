package util_test

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
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

	rsrcInvalid := map[v1.ResourceName]string{
		"cpu":    "1",
		"memory": "2Gi",
		"foo":    "bar",
	}

	actual, err := util.BuildResourceList(rsrcInvalid)
	if err == nil {
		t.Errorf("got: %v\nwant: error", actual)
	}
}

func TestResourceListSum(t *testing.T) {
	r1 := v1.ResourceList{
		"cpu":    resource.MustParse("1"),
		"memory": resource.MustParse("2Gi"),
	}

	r2 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	expected := v1.ResourceList{
		"cpu":            resource.MustParse("3"),
		"memory":         resource.MustParse("6Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	actual := util.ResourceListSum(r1, r2)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	actual = util.ResourceListSum(r2, r1)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestResourceListDiff(t *testing.T) {
	r1 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	r2 := v1.ResourceList{
		"cpu":    resource.MustParse("1"),
		"memory": resource.MustParse("2Gi"),
	}

	expected := v1.ResourceList{
		"cpu":            resource.MustParse("1"),
		"memory":         resource.MustParse("2Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	actual, _ := util.ResourceListDiff(r1, r2)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	actual, err := util.ResourceListDiff(r2, r1)
	if err != util.ErrResourceListDiffNotGE {
		t.Errorf("got: %v\nwant: %v", actual, err)
	}
}

func TestResourceListGE(t *testing.T) {
	r1 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	expected := true
	actual := util.ResourceListGE(r1, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	r2 := v1.ResourceList{
		"cpu":    resource.MustParse("1"),
		"memory": resource.MustParse("2Gi"),
	}

	expected = true
	actual = util.ResourceListGE(r1, r2)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	expected = false
	actual = util.ResourceListGE(r2, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	r3 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("2Gi"),
		"nvidia.com/gpu": resource.MustParse("2"),
	}

	expected = false
	actual = util.ResourceListGE(r1, r3)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	expected = false
	actual = util.ResourceListGE(r3, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
