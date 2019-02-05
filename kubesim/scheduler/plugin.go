package scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// type FitPredicate = func(pod *v1.Pod, meta predicates.PredicateMetadata, nodeInfo *nodeinfo.NodeInfo) (bool, []predicates.PredicateFailureReason, error)

// DummyPredicateMetadata implements predicates.DummyPredicateMetadata
type DummyPredicateMetadata struct{}

func (d *DummyPredicateMetadata) ShallowCopy() predicates.PredicateMetadata             { return d }
func (d *DummyPredicateMetadata) AddPod(pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo) error { return nil }
func (d *DummyPredicateMetadata) RemovePod(pod *v1.Pod) error                           { return nil }

// type PriorityMapFunction = func(pod *v1.Pod, meta interface{}, nodeInfo *nodeinfo.NodeInfo) (api.HostPriority, error)
// type PriorityReduceFunction = func(pod *v1.Pod, meta interface{}, nodeNameToInfo map[string]*nodeinfo.NodeInfo, result api.HostPriorityList) error

type DummyPriorityMetadata struct{}
