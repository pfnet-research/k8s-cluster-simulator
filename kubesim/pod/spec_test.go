package pod

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestParseSpecYAML(t *testing.T) {
	str := `
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

	actual, err := parseSpecYAML(str)
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
	if reflect.DeepEqual(actual[0], expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

	expected = specPhase{
		seconds: 10,
		resourceUsage: v1.ResourceList{
			"cpu":            resource.MustParse("2"),
			"memory":         resource.MustParse("4Gi"),
			"nvidia.com/gpu": resource.MustParse("1"),
		},
	}
	if reflect.DeepEqual(actual[1], expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
