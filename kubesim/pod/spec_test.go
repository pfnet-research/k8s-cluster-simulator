package pod

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func specPhaseNE(sp1, sp2 specPhase) bool {
	return sp1.seconds != sp2.seconds ||
		!reflect.DeepEqual(sp1.resourceUsage, sp2.resourceUsage)
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
		t.Errorf("got: %v\nwant: %v", actual[0], expected)
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
		t.Errorf("got: %v\nwant: %v", actual[1], expected)
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
	actual, err = parseSpecYAML(yamlStrInvalid)
	if err != errInvalidResourceUsageField {
		t.Errorf("got: %v\nwant: errInvalidResourceUsageField", actual)
	}
}
