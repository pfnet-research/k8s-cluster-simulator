package scheduler

import (
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

type Filter interface {
	// Filter filters out nodes that cannot run the pod.
	//
	// Scheduler runs filter plugins per node in the same order that they are registered,
	// but scheduler may run these filter function for multiple nodes in parallel.
	// So these plugins must use synchronization when they modify state.
	Filter(pod *v1.Pod, nodes [](*v1.Node)) (filteredNodes [](*v1.Node), err error)
}

type Scorer interface {
	// Score ranks nodes that have passed the filtering stage.
	//
	// Similar to Filter plugins, these are called per node serially in the same order registered,
	// but scheduler may run them for multiple nodes in parallel.
	//
	// Each one of these functions return a score for the given node.
	// The score is multiplied by the weight of the function and aggregated with the result of
	// other scoring functions to yield a total score for the node.
	//
	// These functions can never block scheduling.
	// In case of an error they should return zero for the Node being ranked.
	Score(pod *v1.Pod, nodes [](*v1.Node)) (hostPriorities *schedulerapi.HostPriorityList, weight int, err error)
}
