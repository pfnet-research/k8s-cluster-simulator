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

package pod

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func specPhaseNE(sp1, sp2 specPhase) bool {
	return sp1.seconds != sp2.seconds ||
		!reflect.DeepEqual(sp1.resourceUsage, sp2.resourceUsage)
}

func TestParseSpec(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	}

	_, err := parseSpec(pod)
	if err == nil {
		t.Error("nil error")
	}

	pod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "default",
			Annotations: map[string]string{
				"simSpec": `
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 1
`,
			},
		},
	}

	actual, err := parseSpec(pod)
	if err != nil {
		t.Errorf("error %s", err.Error())
	}

	expected := specPhase{
		seconds: 5,
		resourceUsage: v1.ResourceList{
			"cpu":            resource.MustParse("1"),
			"memory":         resource.MustParse("2Gi"),
			"nvidia.com/gpu": resource.MustParse("1"),
		},
	}

	if specPhaseNE(expected, actual[0]) {
		t.Errorf("got: %+v\nwant: %+v", actual[0], expected)
	}
}

func TestParseSpecYAML(t *testing.T) {
	yamlStr := `
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10
  resourceUsage:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
`

	actual, err := parseSpecYAML(yamlStr)
	if err != nil {
		t.Errorf("error %s", err.Error())
	}

	expected := specPhase{
		seconds: 5,
		resourceUsage: v1.ResourceList{
			"cpu":            resource.MustParse("1"),
			"memory":         resource.MustParse("2Gi"),
			"nvidia.com/gpu": resource.MustParse("0"),
		},
	}

	if specPhaseNE(expected, actual[0]) {
		t.Errorf("got: %+v\nwant: %+v", actual[0], expected)
	}

	expected = specPhase{
		seconds: 10,
		resourceUsage: v1.ResourceList{
			"cpu":            resource.MustParse("2"),
			"memory":         resource.MustParse("4Gi"),
			"nvidia.com/gpu": resource.MustParse("1"),
		},
	}

	if specPhaseNE(expected, actual[1]) {
		t.Errorf("got: %+v\nwant: %+v", actual[1], expected)
	}

	yamlStrInvalid := `
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10
  resourceUsagi:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
`
	_, err = parseSpecYAML(yamlStrInvalid)
	if err == nil {
		t.Error("nil error")
	}
}
