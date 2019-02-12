package scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/log"
)

// dummyPredicateMetadata implements predicates.DummyPredicateMetadata
type dummyPredicateMetadata struct{}
type dummyPriorityMetadata struct{}

func (d *dummyPredicateMetadata) ShallowCopy() predicates.PredicateMetadata             { return d }
func (d *dummyPredicateMetadata) AddPod(pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo) error { return nil }
func (d *dummyPredicateMetadata) RemovePod(pod *v1.Pod) error                           { return nil }

// type FitPredicate = func(pod *v1.Pod, meta predicates.PredicateMetadata, nodeInfo *nodeinfo.NodeInfo) (bool, []predicates.PredicateFailureReason, error)

func pluginPredicate(
	pred *predicates.FitPredicate,
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeMap map[string]*node.Node,
	failedPredicateMap core.FailedPredicateMap,
	errs errors.MessageCountMap) (filteredNodes []*v1.Node) {
	log.L.Tracef("Predicating nodes %v", nodes)
	for _, node := range nodes {
		fits, failureReason, err := (*pred)(pod, &dummyPredicateMetadata{}, nodeMap[node.Name].ToNodeInfo())
		if err != nil {
			errs[err.Error()]++
		}
		if fits {
			filteredNodes = append(filteredNodes, node)
		} else {
			failedPredicateMap[node.Name] = failureReason
		}
	}

	log.L.Tracef("Predicated nodes %v", filteredNodes)
	return filteredNodes
}

// type PriorityMapFunction = func(pod *v1.Pod, meta interface{}, nodeInfo *nodeinfo.NodeInfo) (api.HostPriority, error)
// type PriorityReduceFunction = func(pod *v1.Pod, meta interface{}, nodeNameToInfo map[string]*nodeinfo.NodeInfo, result api.HostPriorityList) error

func pluginPrioritize(
	prioritizer *priorities.PriorityConfig,
	pod *v1.Pod,
	filteredNodes []*v1.Node,
	nodeMap map[string]*node.Node,
	errs []error) api.HostPriorityList {
	log.L.Tracef("Plugin %q: Prioritizing nodes %v", prioritizer.Name, filteredNodes)

	prios := make(api.HostPriorityList, 0, len(filteredNodes))
	for i, node := range filteredNodes {
		prio, err := prioritizer.Map(pod, &dummyPriorityMetadata{}, nodeMap[node.Name].ToNodeInfo())
		if err != nil {
			errs = append(errs, err)
		}
		prios[i] = prio
	}

	if prioritizer.Reduce != nil {
		nodeInfoMap := map[string]*nodeinfo.NodeInfo{}
		for nodeName, node := range nodeMap {
			nodeInfoMap[nodeName] = node.ToNodeInfo()
		}

		err := prioritizer.Reduce(pod, &dummyPriorityMetadata{}, nodeInfoMap, prios)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for i := range prios {
		prios[i].Score *= prioritizer.Weight
	}

	log.L.Tracef("Prioritized %v", prios)
	return prios
}
