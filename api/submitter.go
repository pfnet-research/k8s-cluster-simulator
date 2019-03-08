package api

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
)

type SubmitterEvent interface {
	IsSubmitterEvent() bool
}

type SubmitEvent struct {
	Pod *v1.Pod
}

type DeleteEvent struct {
	PodName      string
	PodNamespace string
}

type UpdateEvent struct {
	PodName      string
	PodNamespace string
	NewPod       *v1.Pod
}

func (s *SubmitEvent) IsSubmitterEvent() bool { return true }
func (d *DeleteEvent) IsSubmitterEvent() bool { return true }
func (u *UpdateEvent) IsSubmitterEvent() bool { return true }

// Submitter defines the submitter interface.
type Submitter interface {
	// Submitter submits pods to the simulated cluster. They are called in the same order that they
	// are registered.
	// These functions must *not* block.
	Submit(
		clock clock.Clock,
		nodeLister algorithm.NodeLister,
		metrics metrics.Metrics) ([]SubmitterEvent, error)
}
