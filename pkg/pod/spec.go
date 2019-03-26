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
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// spec represents a list of a pod's resource usage spec of each execution phase.
type spec []specPhase

// specPhase represents a pod's resource usage spec of one execution phase.
type specPhase struct {
	seconds       int32
	resourceUsage v1.ResourceList
}

// parseSpec parses the pod's "simSpec" annotation into spec.
// Returns error if the "simSpec" annotation does not exist, or the failed to parse.
func parseSpec(pod *v1.Pod) (spec, error) {
	specAnnot, ok := pod.ObjectMeta.Annotations["simSpec"]
	if !ok {
		return nil, strongerrors.InvalidArgument(errors.Errorf("simSpec annotation not defined"))
	}

	return parseSpecYAML(specAnnot)
}

// parseSpecYAML parses the YAML into spec.
// Returns error if failed to parse.
func parseSpecYAML(specYAML string) (spec, error) {
	type specPhaseYAML struct {
		Seconds       int32                      `yaml:"seconds"`
		ResourceUsage map[v1.ResourceName]string `yaml:"resourceUsage"`
	}

	specUnmarshalled := []specPhaseYAML{}
	if err := yaml.Unmarshal([]byte(specYAML), &specUnmarshalled); err != nil {
		return nil, err
	}

	spec := spec{}
	for _, phase := range specUnmarshalled {
		if phase.ResourceUsage == nil {
			return nil, errors.New("Invalid spec.resoruceUsage field")
		}

		resourceUsage, err := util.BuildResourceList(phase.ResourceUsage)
		if err != nil {
			return nil, err
		}
		spec = append(spec, specPhase{
			seconds:       phase.Seconds,
			resourceUsage: resourceUsage,
		})
	}

	return spec, nil
}
