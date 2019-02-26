package scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/log"
)

// dummyPredicateMetadata implements predicates.PredicateMetadata.
type dummyPredicateMetadata struct{}
type dummyPriorityMetadata struct{}

func (d *dummyPredicateMetadata) ShallowCopy() predicates.PredicateMetadata             { return d }
func (d *dummyPredicateMetadata) AddPod(pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo) error { return nil }
func (d *dummyPredicateMetadata) RemovePod(pod *v1.Pod) error                           { return nil }

func callPredicatePlugin(
	name string,
	pred *predicates.FitPredicate,
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	failedPredicateMap core.FailedPredicateMap,
	errs errors.MessageCountMap) (filteredNodes []*v1.Node) {

	log.L.Tracef("Plugin %s: Predicating nodes %v", name, nodes)

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Plugin %s: Predicating nodes %v", name, nodeNames)

	filteredNodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		fits, failureReason, err := (*pred)(pod, &dummyPredicateMetadata{}, nodeInfoMap[node.Name])
		if err != nil {
			errs[err.Error()]++
		}
		if fits {
			filteredNodes = append(filteredNodes, node)
			filteredNodeNames = append(filteredNodeNames, node.Name)
		} else {
			failedPredicateMap[node.Name] = failureReason
		}
	}

	log.L.Tracef("Plugin %s: Predicated nodes %v", name, filteredNodes)
	log.L.Debugf("Plugin %s: Predicated nodes %v", name, filteredNodeNames)

	return filteredNodes
}

func callPrioritizePlugin(
	prioritizer *priorities.PriorityConfig,
	pod *v1.Pod,
	filteredNodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	errs []error) api.HostPriorityList {

	log.L.Tracef("Plugin %s: Prioritizing nodes %v", prioritizer.Name, filteredNodes)

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Plugin %s: Prioritizing nodes %v", prioritizer.Name, nodeNames)

	prios := make(api.HostPriorityList, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		prio, err := prioritizer.Map(pod, &dummyPriorityMetadata{}, nodeInfoMap[node.Name])
		if err != nil {
			errs = append(errs, err)
		}
		prios = append(prios, prio)
	}

	if prioritizer.Reduce != nil {
		err := prioritizer.Reduce(pod, &dummyPriorityMetadata{}, nodeInfoMap, prios)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for i := range prios {
		prios[i].Score *= prioritizer.Weight
	}

	log.L.Debugf("Plugin %s: Prioritized %v", prioritizer.Name, prios)

	return prios
}
