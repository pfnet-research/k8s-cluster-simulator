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

package util_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/scheduling"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

func resourceListEq(r1, r2 v1.ResourceList) bool {
	helper := func(r1, r2 v1.ResourceList) bool {
		for r1Key, r1Val := range r1 {
			r2Val, ok := r2[r1Key]
			if !ok || r1Val.Value() != r2Val.Value() {
				return false
			}
		}
		return true
	}

	return helper(r1, r2) && helper(r2, r1)
}

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
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	rsrcInvalid := map[v1.ResourceName]string{
		"cpu":    "1",
		"memory": "2Gi",
		"foo":    "bar",
	}

	_, err := util.BuildResourceList(rsrcInvalid)
	assert.EqualError(t, err, "invalid foo value \"bar\"")
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
	if !resourceListEq(expected, actual) {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	actual = util.ResourceListSum(r2, r1)
	if !resourceListEq(expected, actual) {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestPodTotalResourceRequests(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
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
				{
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
	actual := util.PodTotalResourceRequests(&pod)

	if !resourceListEq(expected, actual) {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
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
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	r2 := v1.ResourceList{
		"cpu":    resource.MustParse("1"),
		"memory": resource.MustParse("2Gi"),
	}

	expected = true
	actual = util.ResourceListGE(r1, r2)
	if expected != actual {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	expected = false
	actual = util.ResourceListGE(r2, r1)
	if expected != actual {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	r3 := v1.ResourceList{
		"cpu":            resource.MustParse("2"),
		"memory":         resource.MustParse("2Gi"),
		"nvidia.com/gpu": resource.MustParse("2"),
	}

	expected = false
	actual = util.ResourceListGE(r1, r3)
	if expected != actual {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	expected = false
	actual = util.ResourceListGE(r3, r1)
	if expected != actual {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestPodPriority(t *testing.T) {
	expected := int32(9)
	actual := util.PodPriority(&v1.Pod{
		Spec: v1.PodSpec{
			Priority: &expected,
		},
	})

	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	actual = util.PodPriority(&v1.Pod{})
	expected = int32(scheduling.DefaultPriorityWhenNoDefaultClassExists)

	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestPodKey(t *testing.T) {
	actual, _ := util.PodKey(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace-0",
			Name:      "name-0",
		},
	})
	expected := "namespace-0/name-0"

	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	_, err := util.PodKey(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "",
			Name:      "name-0",
		},
	})
	assert.EqualError(t, err, "Empty pod namespace")

	_, err = util.PodKey(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace-0",
			Name:      "",
		},
	})
	assert.EqualError(t, err, "Empty pod name")
}
