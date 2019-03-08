package submitter

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
)

// Submitter defines the submitter interface.
type Submitter interface {
	// Submitter submits pods to the simulated cluster. They are called in the same order that they
	// are registered.
	// These functions must *not* block.
	Submit(
		clock clock.Clock,
		nodeLister algorithm.NodeLister,
		metrics metrics.Metrics) ([]Event, error)
}

// Event defines the interface of a submitter event.
type Event interface {
	IsSubmitterEvent() bool
}

// SubmitEvent represents an event of submitting a pod to the cluster.
type SubmitEvent struct {
	Pod *v1.Pod
}

// DeleteEvent represents an event of deleting a pending or running pod from the cluster.
type DeleteEvent struct {
	PodName      string
	PodNamespace string
}

// UpdateEvent represents an event of updating the manifest of a pending pod.
type UpdateEvent struct {
	PodName      string
	PodNamespace string
	NewPod       *v1.Pod
}

func (s *SubmitEvent) IsSubmitterEvent() bool { return true }
func (d *DeleteEvent) IsSubmitterEvent() bool { return true }
func (u *UpdateEvent) IsSubmitterEvent() bool { return true }
