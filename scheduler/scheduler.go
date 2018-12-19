package scheduler

import (
	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

type Submitter interface {
	Submit(clock clock.Clock, nodes []*v1.Node) (pods []*v1.Pod, err error)
}

// Filter plugin interface
type Filter interface {
	// Filter filters out nodes that cannot run the pod.
	//
	// Scheduler runs filter plugins per node in the same order that they are registered,
	// but scheduler may run these filter function for multiple nodes in parallel.
	// So these plugins must use synchronization when they modify state.
	//
	// Scheduler stops running the remaining filter functions for a node once one of these filters
	// fails for the node.
	Filter(pod *v1.Pod, node *v1.Node) (ok bool, err error)
}

// NodeScore represents the score of scheduling to a particular node.
// Higher score means higher priority.
type NodeScore struct {
	// Name of the node.
	Node string
	// Score associated with the node.
	Score int
}

// Scorer plugin interface
type Scorer interface {
	// Score ranks nodes that have passed the filtering stage.
	//
	// Similar to Filter plugins, these are called per node serially in the same order registered,
	// but scheduler may run them for multiple nodes in parallel.
	//
	// Each one of these functions return a score for the given node.
	// The score is multiplied by the weight of the function and aggregated with the result of
	// other scoring functions to yield a total score for the node.
	// Both the score and weight must be positive values.
	//
	// These functions can never block scheduling.
	// In case of an error they should return zero for the Node being ranked.
	Score(pod *v1.Pod, nodes []*v1.Node) (scores []NodeScore, weight int, err error)
}
