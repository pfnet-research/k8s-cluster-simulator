package sim

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

// simPod represents a simulated pod.
type simPod struct {
	pod        *v1.Pod
	startClock Time
	status     simPodStatus
	spec       simSpec
}

type simPodStatus int

const (
	simPodOk           simPodStatus = iota // pod is successfully scheduled to a node
	simPodOverCapacity                     // pod is failed to start due to capacity over
)

// passedSeconds returns elapsed duration after the pod started in seconds.
func (pod *simPod) passedSeconds(clock Time) int32 {
	return int32(clock.Sub(pod.startClock).Seconds())
}

// resourceUsage returns resource usage of the pod at the time it has run passedSeconds duration.
func (pod *simPod) resourceUsage(clock Time) v1.ResourceList {
	passedSeconds := pod.passedSeconds(clock)
	phaseSecondsAcc := int32(0)
	for _, phase := range pod.spec {
		if passedSeconds < phaseSecondsAcc+phase.seconds {
			return phase.resourceUsage
		}
		phaseSecondsAcc += phase.seconds
	}
	return v1.ResourceList{}
}

// totalSeconds returns total duration of the pod in seconds.
func (pod *simPod) totalSeconds() int32 {
	phaseSecondsTotal := int32(0)
	for _, phase := range pod.spec {
		phaseSecondsTotal += phase.seconds
	}
	return phaseSecondsTotal
}

func (pod *simPod) finishClock() Time {
	return pod.startClock.Add(time.Duration(pod.totalSeconds()) * time.Second)
}

func (pod *simPod) isTerminated(clock Time) bool {
	return pod.passedSeconds(clock) >= pod.totalSeconds()
}

func (pod *simPod) buildStatus(clock Time) v1.PodStatus {
	var status v1.PodStatus

	switch pod.status {
	case simPodOverCapacity:
		status = v1.PodStatus{
			Phase:   v1.PodFailed,
			Reason:  "CapacityExceeded",
			Message: "Pod cannot be started due to exceeded capacity",
		}
	case simPodOk:
		startTime := pod.startClock.ToMetaV1()
		finishTime := pod.finishClock().ToMetaV1()

		var phase v1.PodPhase
		var containerState v1.ContainerState

		if pod.isTerminated(clock) {
			phase = v1.PodSucceeded
			containerState = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					ExitCode:   0,
					Reason:     "Succeeded", // TODO
					Message:    "All containers in the pod have voluntarily terminated",
					StartedAt:  startTime,
					FinishedAt: finishTime,
				}}
		} else {
			phase = v1.PodRunning
			containerState = v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: startTime,
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

		for _, container := range pod.pod.Spec.Containers {
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

// simSpec represents pod's resource usage spec of each execution phase
type simSpec []simSpecPhase

type simSpecPhase struct {
	seconds       int32
	resourceUsage v1.ResourceList
}

// parseSimSpec parses the pod's "spec" annotation into simSpec.
func parseSimSpec(pod *v1.Pod) (simSpec, error) {
	type simSpecPhaseJSON struct {
		Seconds       int32           `json:"seconds"`
		ResourceUsage v1.ResourceList `json:"resourceUsage"`
	}

	simSpecAnnot, ok := pod.ObjectMeta.Annotations["spec"]
	if !ok {
		return nil, strongerrors.InvalidArgument(errors.Errorf("spec not defined"))
	}

	simSpecJSON := []simSpecPhaseJSON{}
	err := json.Unmarshal([](byte)(simSpecAnnot), &simSpecJSON)
	if err != nil {
		return nil, err
	}

	spec := simSpec{}
	for _, phase := range simSpecJSON {
		spec = append(spec, simSpecPhase{
			seconds:       phase.Seconds,
			resourceUsage: phase.ResourceUsage,
		})
	}

	return spec, nil
}

// podMap stores a map associating "key" with *v1.Pod.
// It wraps sync.Map for type-safety.
type podMap struct {
	sync.Map
}

func (m *podMap) load(key string) (simPod, bool) {
	pod, ok := m.Load(key)
	if !ok {
		return simPod{}, false
	}
	return pod.(simPod), true
}

func (m *podMap) store(key string, pod simPod) {
	m.Store(key, pod)
}

func (m *podMap) remove(key string) {
	m.Delete(key)
}

// listPods returns an array of pods
func (m *podMap) listPods() []*v1.Pod {
	pods := []*v1.Pod{}
	m.foreach(func(_ string, pod simPod) bool {
		pods = append(pods, pod.pod)
		return true
	})
	return pods
}

// foreach applies a function to each pair of key and pod
func (m *podMap) foreach(f func(string, simPod) bool) {
	g := func(key, pod interface{}) bool {
		return f(key.(string), pod.(simPod))
	}
	m.Range(g)
}
