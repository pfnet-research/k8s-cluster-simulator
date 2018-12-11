package pod

import (
	"time"

	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

// Pod represents a simulated pod.
type Pod struct {
	pod        *v1.Pod
	spec       spec
	startClock clock.Clock
	status     Status
}

// Status represents status of a pod.
type Status int

const (
	// Ok means that the pod is successfully scheduled to a node
	Ok Status = iota
	// OverCapacity means that the pod is failed to start due to capacity over
	OverCapacity
)

// NewPod creates a pod with the pod definition, the starting time, and the status.
// Returns error if it fails to parse the simulation spec of the pod.
func NewPod(pod *v1.Pod, startClock clock.Clock, status Status) (*Pod, error) {
	spec, err := parseSpec(pod)
	if err != nil {
		return nil, err
	}

	p := Pod{pod, spec, startClock, status}
	return &p, nil
}

// ToV1 returns definition of this pod.
func (pod *Pod) ToV1() *v1.Pod {
	return pod.pod
}

// ResourceUsage returns resource usage of the pod at the time clock.
func (pod *Pod) ResourceUsage(clock clock.Clock) v1.ResourceList {
	if !pod.IsRunning(clock) {
		return v1.ResourceList{}
	}

	passedSeconds := pod.passedSeconds(clock)
	phaseSecondsAcc := int32(0)
	for _, phase := range pod.spec {
		if passedSeconds < phaseSecondsAcc+phase.seconds {
			return phase.resourceUsage
		}
		phaseSecondsAcc += phase.seconds
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
	var status v1.PodStatus

	switch pod.status {
	case OverCapacity:
		status = v1.PodStatus{
			Phase:   v1.PodFailed,
			Reason:  "CapacityExceeded",
			Message: "Pod cannot be started due to exceeded capacity",
		}
	case Ok:
		startTime := pod.startClock.ToMetaV1()
		finishTime := pod.finishClock().ToMetaV1()

		var phase v1.PodPhase
		var containerState v1.ContainerState
		if pod.IsRunning(clock) {
			phase = v1.PodRunning
			containerState = v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: startTime,
				}}
		} else {
			phase = v1.PodSucceeded
			containerState = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					ExitCode:   0,
					Reason:     "Succeeded",
					Message:    "All containers in the pod have voluntarily terminated",
					StartedAt:  startTime,
					FinishedAt: finishTime,
				}}
		}

		status = v1.PodStatus{
			Phase:     phase,
			HostIP:    "1.2.3.4",
			PodIP:     "5.6.7.8",
			StartTime: &startTime,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodInitialized,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionTrue,
				},
			},
		}

		for _, container := range pod.ToV1().Spec.Containers {
			status.ContainerStatuses = append(status.ContainerStatuses, v1.ContainerStatus{
				Name:         container.Name,
				Image:        container.Image,
				Ready:        true,
				RestartCount: 0,
				State:        containerState,
			})
		}
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

func (pod *Pod) finishClock() clock.Clock {
	return pod.startClock.Add(time.Duration(pod.totalSeconds()) * time.Second)
}
