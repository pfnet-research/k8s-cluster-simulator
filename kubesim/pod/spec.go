package pod

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// spec represents pod's resource usage spec of each execution phase.
type spec []specPhase

// specPhase represents a pod's resource usage spec in an execution phase.
type specPhase struct {
	seconds       int32
	resourceUsage v1.ResourceList
}

// parseSpec parses the pod's "spec" annotation into spec.
func parseSpec(pod *v1.Pod) (spec, error) {
	specAnnot, ok := pod.ObjectMeta.Annotations["simSpec"]
	if !ok {
		return nil, strongerrors.InvalidArgument(errors.Errorf("simSpec annotation not defined"))
	}

	return parseSpecYAML(specAnnot)
}

// parseSpec parses the YAML into spec.
func parseSpecYAML(specYAML string) (spec, error) {
	type specPhaseYAML struct {
		Seconds       int32
		ResourceUsage map[v1.ResourceName]string
	}

	specUnmarshalled := []specPhaseYAML{}
	if err := yaml.Unmarshal([](byte)(specYAML), &specUnmarshalled); err != nil {
		return nil, err
	}

	spec := spec{}
	for _, phase := range specUnmarshalled {
		resourceUsage, err := util.BuildResourceList(phase.ResourceUsage)
		if err != nil {
			return spec, err
		}
		spec = append(spec, specPhase{
			seconds:       phase.Seconds,
			resourceUsage: resourceUsage,
		})
	}

	return spec, nil
}
