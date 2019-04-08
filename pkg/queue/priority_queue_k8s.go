/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Modifications copyright 2019 Preferred Networks, Inc.
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

// podTimestamp function was copied from
// k8s.io/kubernetes/pkg/scheduler/internal/queue/scheduling_queue.go by the authors of
// k8s-cluster-simulator, and modified so that it would be compatible with k8s-cluster-simulator.

package queue

import (
	v1 "k8s.io/api/core/v1"
	v1pod "k8s.io/kubernetes/pkg/api/v1/pod"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
)

func podTimestamp(pod *v1.Pod) clock.Clock {
	_, condition := v1pod.GetPodCondition(&pod.Status, v1.PodScheduled)
	if condition == nil {
		return clock.NewClockWithMetaV1(pod.CreationTimestamp)
	}

	if condition.LastProbeTime.IsZero() {
		return clock.NewClockWithMetaV1(condition.LastTransitionTime)
	}
	return clock.NewClockWithMetaV1(condition.LastProbeTime)
}
