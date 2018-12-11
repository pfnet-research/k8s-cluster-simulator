package pod

import (
	"encoding/json"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

// spec represents pod's resource usage spec of each execution phase.
type spec []specPhase

// specPhase represents a pod's resource usage spec in an execution phase.
type specPhase struct {
	seconds       int32
	resourceUsage v1.ResourceList
}

// parseSpec parses the pod's "spec" annotation into Spec.
func parseSpec(pod *v1.Pod) (spec, error) {
	type specPhaseJSON struct {
		Seconds       int32           `json:"seconds"`
		ResourceUsage v1.ResourceList `json:"resourceUsage"`
	}

	specAnnot, ok := pod.ObjectMeta.Annotations["spec"]
	if !ok {
		return nil, strongerrors.InvalidArgument(errors.Errorf("spec not defined"))
	}

	specJSON := []specPhaseJSON{}
	if err := json.Unmarshal([](byte)(specAnnot), &specJSON); err != nil {
		return nil, err
	}

	spec := spec{}
	for _, phase := range specJSON {
		spec = append(spec, specPhase{
			seconds:       phase.Seconds,
			resourceUsage: phase.ResourceUsage,
		})
	}

	return spec, nil
}
