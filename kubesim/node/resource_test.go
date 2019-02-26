package node

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func resourceListNE(r1, r2 v1.ResourceList) bool {
	if len(r1) != len(r2) {
		return true
	}
	for r1Key, r1Val := range r1 {
		r2Val := r2[r1Key]
		if r1Val.Value() != r2Val.Value() {
			return true
		}
	}

	return false
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

	actual := resourceListSum(r1, r2)
	if resourceListNE(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	actual = resourceListSum(r2, r1)
	if resourceListNE(expected, actual) {
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

	actual, _ := resourceListDiff(r1, r2)
	if resourceListNE(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	actual, err := resourceListDiff(r2, r1)
	if err != errResourceListDiffNotGE {
		t.Errorf("got: %v\nwant: %v", actual, err)
	}
}

func TestExtraceResourceList(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse("3"),
							"memory":         resource.MustParse("5Gi"),
							"nvidia.com/gpu": resource.MustParse("1"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("4"),
							"memory":         resource.MustParse("6Gi"),
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
				v1.Container{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse("2"),
							"memory":         resource.MustParse("4Gi"),
							"nvidia.com/gpu": resource.MustParse("2"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("3"),
							"memory":         resource.MustParse("5Gi"),
							"nvidia.com/gpu": resource.MustParse("3"),
						},
					},
				},
			},
		},
	}

	expected := v1.ResourceList{
		"cpu":            resource.MustParse("5"),
		"memory":         resource.MustParse("9Gi"),
		"nvidia.com/gpu": resource.MustParse("3"),
	}
	actual := extractResourceRequest(&pod)

	if resourceListNE(expected, actual) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestResourceListGE(t *testing.T) {
	r1 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("4Gi"),
		"nvidia.com/gpu": resource.MustParse("1"),
	}

	expected := true
	actual := resourceListGE(r1, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	r2 := v1.ResourceList{
		"cpu":    resource.MustParse("1"),
		"memory": resource.MustParse("2Gi"),
	}

	expected = true
	actual = resourceListGE(r1, r2)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	expected = false
	actual = resourceListGE(r2, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	r3 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("2Gi"),
		"nvidia.com/gpu": resource.MustParse("2"),
	}

	expected = false
	actual = resourceListGE(r1, r3)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	expected = false
	actual = resourceListGE(r3, r1)
	if expected != actual {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
