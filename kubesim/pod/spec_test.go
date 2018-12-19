package pod

import (
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func specPhaseNE(sp1, sp2 specPhase) bool {
	if sp1.seconds != sp2.seconds {
		return true
	}

	r1 := sp1.resourceUsage
	r2 := sp2.resourceUsage

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
}
