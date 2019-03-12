package scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/k8s-cluster-simulator/kubesim/clock"
	"github.com/ordovicia/k8s-cluster-simulator/kubesim/queue"
)

// Scheduler defines the lowest-level scheduler interface.
type Scheduler interface {
	// Schedule makes scheduling decisions for (subset of) pending pods and running pods.
	// The return value is a list of scheduling events.
	Schedule(
		clock clock.Clock,
		podQueue queue.PodQueue,
		nodeLister algorithm.NodeLister,
		nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error)
}

// Event defines the interface of a scheduling event.
type Event interface {
	IsSchedulerEvent() bool
}

// BindEvent represents an event of deciding the binding of a pod to a node.
type BindEvent struct {
	Pod            *v1.Pod
	ScheduleResult core.ScheduleResult
}

// DeleteEvent represents an event of the deleting a bound pod on a node.
type DeleteEvent struct {
	PodNamespace string
	PodName      string
	NodeName     string
}

func (b *BindEvent) IsSchedulerEvent() bool   { return true }
func (d *DeleteEvent) IsSchedulerEvent() bool { return true }
