package api

import (
	v1 "k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

// Submitter interface
type Submitter interface {
	// Submitter submits pods to the simulated cluster.
	// They are called in the same order that they are registered.
	//
	// These functions must not block the main loop of the simulator.
	Submit(clock clock.Clock, nodes []*v1.Node) (pods []*v1.Pod, err error)
}
