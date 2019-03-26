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
	"encoding/json"
	"time"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// Pod represents a simulated pod.
type Pod struct {
	v1      *v1.Pod
	spec    spec
	boundAt clock.Clock
	status  Status
	node    string
}

// Metrics is a metrics of a pod at one time point.
type Metrics struct {
	ResourceRequest v1.ResourceList
	ResourceLimit   v1.ResourceList
	ResourceUsage   v1.ResourceList

	BoundAt         clock.Clock
	Node            string
	ExecutedSeconds int32

	Priority int32
	Status   Status
}

// Status represents status of a Pod.
type Status int

const (
	// Ok indicates that the pod has been successfully started on a node.
	// Whether the pod is running or has spontaneously terminated is determined by its total
	// execution time and the clock.
	Ok Status = iota

	// Deleted indicates that the pod has been deleted from the cluster.
	// Whether the pod is terminating (i.e., in its grace period) or has been deleted is determined
	// by the length of its grace period and the clock.
	Deleted

	// OverCapacity indicates that the pod failed to start due to over capacity.
	OverCapacity
)

// String implements Stringer interface.
func (status Status) String() string {
	switch status {
	case Ok:
		return "Ok"
	case Deleted:
		return "Deleted"
	case OverCapacity:
		return "OverCapacity"
	default:
		log.L.Panic("Unknown pod.Status")
		return ""
	}
}

// MarshalJSON implements json.Marshaler interface.
func (status Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// NewPod creates a pod with the given v1.Pod, the clock at which the pod was bound to a node, and
// the pod's status.
// Returns error if fails to parse the simulation spec of the pod.
func NewPod(pod *v1.Pod, boundAt clock.Clock, status Status, node string) (*Pod, error) {
	spec, err := parseSpec(pod)
	if err != nil {
		return nil, err
	}

	return &Pod{
		v1:      pod,
		spec:    spec,
		boundAt: boundAt,
		status:  status,
		node:    node,
	}, nil
}

// ToV1 returns v1.Pod representation of this Pod.
func (pod *Pod) ToV1() *v1.Pod {
	return pod.v1
}

// Metrics returns the Metrics of this Pod at the given clock.
func (pod *Pod) Metrics(clock clock.Clock) Metrics {
	return Metrics{
		ResourceRequest: pod.TotalResourceRequests(),
		ResourceLimit:   pod.TotalResourceLimits(),
		ResourceUsage:   pod.ResourceUsage(clock),

		BoundAt:         pod.boundAt,
		Node:            pod.node,
		ExecutedSeconds: int32(pod.executedDuration(clock).Seconds()),

		Priority: util.PodPriority(pod.ToV1()),
		Status:   pod.status,
	}
}

// TotalResourceRequests extracts the total amount of resource requested by this Pod.
func (pod *Pod) TotalResourceRequests() v1.ResourceList {
	return util.PodTotalResourceRequests(pod.ToV1())
}

// TotalResourceLimits extracts the total amount of resource limits of this Pod.
func (pod *Pod) TotalResourceLimits() v1.ResourceList {
	result := v1.ResourceList{}
	for _, container := range pod.ToV1().Spec.Containers {
		result = util.ResourceListSum(result, container.Resources.Limits)
	}
	return result
}

// ResourceUsage returns resource usage of this Pod at the given clock.
func (pod *Pod) ResourceUsage(clock clock.Clock) v1.ResourceList {
	if !(pod.IsRunning(clock) || pod.IsTerminating(clock)) {
		// pod is not using resource
		return v1.ResourceList{}
	}

	executedSeconds := int32(pod.executedDuration(clock).Seconds())
	phaseDurationAcc := int32(0)
	for _, phase := range pod.spec {
		phaseDurationAcc += phase.seconds
		if executedSeconds < phaseDurationAcc {
			return phase.resourceUsage
		}
	}

	log.L.Panic("Unreachable code in pod.ResourceUsage()")
	return v1.ResourceList{}
}

// IsRunning returns whether this Pod is running at the given clock.
// Returns false if this Pod has failed to start.
func (pod *Pod) IsRunning(clock clock.Clock) bool {
	return pod.status == Ok && pod.executedDuration(clock) < pod.totalExecutionDuration()
}

// IsTerminated returns whether this Pod is terminated at the clock.
// If this Pod failed to start, false is returned.
func (pod *Pod) IsTerminated(clock clock.Clock) bool {
	return pod.status == Ok && pod.executedDuration(clock) >= pod.totalExecutionDuration()
}

// IsTerminating returns whether this Pod is terminating (i.e. in its grace period).
func (pod *Pod) IsTerminating(clock clock.Clock) bool {
	return pod.status == Deleted && !pod.IsDeleted(clock)
}

// IsDeleted returns whether this Pod has been deleted.
func (pod *Pod) IsDeleted(clk clock.Clock) bool {
	gp := int64(v1.DefaultTerminationGracePeriodSeconds)
	if pod.v1.Spec.TerminationGracePeriodSeconds != nil {
		gp = *pod.v1.Spec.TerminationGracePeriodSeconds
	}

	return pod.status == Deleted &&
		clk.Sub(clock.NewClockWithMetaV1(*pod.ToV1().DeletionTimestamp)) >= time.Duration(gp)*time.Second
}

// Delete starts to delete this Pod.
func (pod *Pod) Delete(clock clock.Clock) {
	if pod.IsTerminated(clock) || pod.status == Deleted {
		return
	}

	// Running or OverCapacity

	pod.status = Deleted
	deletedAt := clock.ToMetaV1()
	pod.ToV1().DeletionTimestamp = &deletedAt
}

// HasFailedToStart returns whether this Pod has failed to start to a node.
func (pod *Pod) HasFailedToStart() bool {
	return pod.status == OverCapacity
}

// BuildStatus builds a status of this Pod at the given clock, assuming that this Pod has not been
// deleted (but it can be terminating).
func (pod *Pod) BuildStatus(clock clock.Clock) v1.PodStatus {
	status := pod.ToV1().Status

	switch pod.status {
	case OverCapacity:
		status.Phase = v1.PodFailed
		// status.Conditions =
		status.Reason = "CapacityExceeded"
		status.Message = "Pod cannot be started due to the requested resource exceeds the capacity"
	case Ok, Deleted:
		startTime := pod.boundAt.ToMetaV1()
		status.StartTime = &startTime

		var containerState v1.ContainerState
		if pod.IsRunning(clock) || pod.IsTerminating(clock) {
			status.Phase = v1.PodRunning
			containerState = v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: startTime,
				}}
		} else {
			status.Phase = v1.PodSucceeded
			containerState = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					ExitCode: 0,
					// Signal:
					Reason:     "Succeeded",
					Message:    "All containers in the pod have voluntarily terminated",
					StartedAt:  startTime,
					FinishedAt: pod.finishAt().ToMetaV1(),
					// ContainerID:
				}}
		}

		for _, conditionType := range []v1.PodConditionType{v1.PodInitialized, v1.PodReady} {
			util.UpdatePodCondition(clock, &status, &v1.PodCondition{
				Type:               conditionType,
				Status:             v1.ConditionTrue,
				LastProbeTime:      clock.ToMetaV1(),
				LastTransitionTime: startTime,
				// Reason:
				// Message:
			})
		}

		containerStatuses := make([]v1.ContainerStatus, 0, len(pod.ToV1().Spec.Containers))
		for _, container := range pod.ToV1().Spec.Containers {
			containerStatuses = append(containerStatuses, v1.ContainerStatus{
				Name:  container.Name,
				State: containerState,
				// LastTerminationState:
				Ready:        true,
				RestartCount: 0,
				Image:        container.Image,
				// ImageId:
				// ContainerID:
			})
		}

		status.ContainerStatuses = containerStatuses
	}

	return status
}

// executedDuration returns the elapsed duration after this Pod started.
// Returns 0 if the pod failed to start.
func (pod *Pod) executedDuration(clock clock.Clock) time.Duration {
	switch pod.status {
	case Ok:
		elapsed := clock.Sub(pod.boundAt)
		total := pod.totalExecutionDuration()
		if elapsed < total {
			return elapsed
		}
		return total
	case Deleted:
		return pod.ToV1().DeletionTimestamp.Sub(pod.boundAt.ToMetaV1().Time)
	default:
		return 0
	}
}

// totalExecutionDuration returns the total execution duration of this Pod.
func (pod *Pod) totalExecutionDuration() time.Duration {
	phaseSecondsTotal := int32(0)
	for _, phase := range pod.spec {
		phaseSecondsTotal += phase.seconds
	}
	return time.Duration(phaseSecondsTotal) * time.Second
}

// finishAt returns the clock at which this Pod will finish spontaneously.
func (pod *Pod) finishAt() clock.Clock {
	return pod.boundAt.Add(pod.totalExecutionDuration())
}
