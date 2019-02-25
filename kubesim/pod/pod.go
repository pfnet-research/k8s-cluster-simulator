package pod

import (
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/util"
)

// Pod represents a simulated pod.
type Pod struct {
	v1         *v1.Pod
	spec       spec
	startClock clock.Clock
	status     Status
}

// Status represents status of a Pod.
type Status int

const (
	// Ok indicates that the Pod is successfully scheduled to a node.
	Ok Status = iota
	// OverCapacity indicates that the Pod is failed to start due to capacity over.
	OverCapacity
)

// NewPod creates a pod with the v1.Pod definition, the starting time, and the status.
// Returns error if it fails to parse the simulation spec of the pod.
func NewPod(pod *v1.Pod, startClock clock.Clock, status Status) (*Pod, error) {
	spec, err := parseSpec(pod)
	if err != nil {
		return nil, err
	}

	p := Pod{pod, spec, startClock, status}
	return &p, nil
}

// ToV1 returns v1.Pod representation of this pod.
func (pod *Pod) ToV1() *v1.Pod {
	return pod.v1
}

// ResourceUsage returns resource usage of the pod at the clock.
func (pod *Pod) ResourceUsage(clock clock.Clock) v1.ResourceList {
	if !pod.IsRunning(clock) {
		return v1.ResourceList{}
	}

	passedSeconds := pod.passedSeconds(clock)
	phaseSecondsAcc := int32(0)
	for _, phase := range pod.spec {
		phaseSecondsAcc += phase.seconds
		if passedSeconds < phaseSecondsAcc {
			return phase.resourceUsage
		}
	}

	// unreachable
	return v1.ResourceList{}
}

// IsRunning returns whether the pod is running at the clock.
// If the pod failed to be scheduled, false is returned.
func (pod *Pod) IsRunning(clock clock.Clock) bool {
	return pod.status == Ok && pod.passedSeconds(clock) < pod.totalSeconds()
}

// IsTerminated returns whether the pod is terminated at the clock.
// If the pod failed to be scheduled, false is returned.
func (pod *Pod) IsTerminated(clock clock.Clock) bool {
	return pod.status == Ok && pod.passedSeconds(clock) >= pod.totalSeconds()
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
		startTime := pod.startClock.ToMetaV1()
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

// passedSeconds returns elapsed duration in seconds after the pod started.
func (pod *Pod) passedSeconds(clock clock.Clock) int32 {
	if pod.status != Ok {
		return 0
	}
	return int32(clock.Sub(pod.startClock).Seconds())
}

// totalSeconds returns total duration of the pod in seconds.
func (pod *Pod) totalSeconds() int32 {
	phaseSecondsTotal := int32(0)
	for _, phase := range pod.spec {
		phaseSecondsTotal += phase.seconds
	}
	return phaseSecondsTotal
}

// finishClock returns the clock at which this pod finishes.
func (pod *Pod) finishClock() clock.Clock {
	return pod.startClock.Add(time.Duration(pod.totalSeconds()) * time.Second)
}
