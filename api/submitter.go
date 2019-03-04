package api

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/metrics"
)

// Submitter interface
type Submitter interface {
	// Submitter submits pods to the simulated cluster.
	// They are called in the same order that they are registered.
	//
	// These functions must not block the main loop of the simulator.
	Submit(
		clock clock.Clock,
		nodeLister algorithm.NodeLister,
		metrics metrics.Metrics) (pods []*v1.Pod, err error)
}
