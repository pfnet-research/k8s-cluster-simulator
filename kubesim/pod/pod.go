package pod

import (
	"encoding/json"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// Pod represents a simulated pod.
type Pod struct {
	v1      *v1.Pod
	spec    spec
	boundAt clock.Clock
	status  Status
	node    string
}

// Metrics is a metrics of a pod at a time instance.
type Metrics struct {
	ResourceRequest v1.ResourceList
	ResourceLimit   v1.ResourceList
	ResourceUsage   v1.ResourceList

	BoundAt         clock.Clock
	Node            string
	ExecutedSeconds int32

	Status Status
}

// Status represents status of a Pod.
type Status int

const (
	// Ok indicates that the pod is successfully bound to a node.
	Ok Status = iota
	// OverCapacity indicates that the pod is failed to start due to capacity over.
	OverCapacity
)

// MarshalJSON implements json.Marshaler.
func (status Status) MarshalJSON() ([]byte, error) {
	var s string
	switch status {
	case Ok:
		s = "Ok"
	case OverCapacity:
		s = "OverCapacity"
	}
	return json.Marshal(s)
}

// NewPod creates a pod with the v1.Pod definition, the starting time, and the status.
// Returns error if it fails to parse the simulation spec of the pod.
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

// ToV1 returns v1.Pod representation of this pod.
func (pod *Pod) ToV1() *v1.Pod {
	return pod.v1
}

// Metrics returns the Metrics at the time clock.
func (pod *Pod) Metrics(clock clock.Clock) Metrics {
	return Metrics{
		ResourceRequest: pod.TotalResourceRequests(),
		ResourceLimit:   pod.TotalResourceLimits(),
		ResourceUsage:   pod.ResourceUsage(clock),

		BoundAt:         pod.boundAt,
		Node:            pod.node,
		ExecutedSeconds: int32(pod.executedDuration(clock).Seconds()),

		Status: pod.status,
	}
}

// TotalResourceRequests extracts the total amount of resource requested by this pod.
func (pod *Pod) TotalResourceRequests() v1.ResourceList {
	return util.PodTotalResourceRequests(pod.ToV1())
}

// TotalResourceLimits extracts the total of resource limits of this pod.
func (pod *Pod) TotalResourceLimits() v1.ResourceList {
	result := v1.ResourceList{}
	for _, container := range pod.ToV1().Spec.Containers {
		result = util.ResourceListSum(result, container.Resources.Limits)
	}
	return result
}

// ResourceUsage returns resource usage of the pod at the clock.
func (pod *Pod) ResourceUsage(clock clock.Clock) v1.ResourceList {
	if !pod.IsRunning(clock) {
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

	// unreachable
	return v1.ResourceList{}
}

// IsRunning returns whether the pod is running at the clock.
// If the pod failed to be bound, false is returned.
func (pod *Pod) IsRunning(clock clock.Clock) bool {
	return pod.status == Ok && pod.executedDuration(clock) < pod.totalExecutionDuration()
}

// IsTerminated returns whether the pod is terminated at the clock.
// If the pod failed to be bound, false is returned.
func (pod *Pod) IsTerminated(clock clock.Clock) bool {
	return pod.status == Ok && pod.executedDuration(clock) >= pod.totalExecutionDuration()
}

// IsBindingFailed returns whether the pod failed to be bound to a node.
func (pod *Pod) IsBindingFailed() bool {
	return pod.status != Ok
}

// BuildStatus builds a status of this pod at the clock.
func (pod *Pod) BuildStatus(clock clock.Clock) v1.PodStatus {
	status := pod.ToV1().Status

	switch pod.status {
	case OverCapacity:
		status.Phase = v1.PodFailed
		// status.Conditions
		status.Reason = "CapacityExceeded"
		status.Message = "Pod cannot be started due to the requested resource exceeds the capacity"
	case Ok:
		startTime := pod.boundAt.ToMetaV1()
		finishTime := pod.finishClock().ToMetaV1()

		status.StartTime = &startTime

		var containerState v1.ContainerState
		if pod.IsRunning(clock) {
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
					FinishedAt: finishTime,
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

// executedDuration returns elapsed duration after the pod started.
// Returns 0 if the pod failed to be bound.
func (pod *Pod) executedDuration(clock clock.Clock) time.Duration {
	if pod.status != Ok {
		return 0
	}
	return clock.Sub(pod.boundAt)
}

// totalExecutionDuration returns total duration of the pod.
func (pod *Pod) totalExecutionDuration() time.Duration {
	phaseSecondsTotal := int32(0)
	for _, phase := range pod.spec {
		phaseSecondsTotal += phase.seconds
	}
	return time.Duration(phaseSecondsTotal) * time.Second
}

// finishClock returns the clock at which this pod finishes.
func (pod *Pod) finishClock() clock.Clock {
	return pod.boundAt.Add(pod.totalExecutionDuration())
}
