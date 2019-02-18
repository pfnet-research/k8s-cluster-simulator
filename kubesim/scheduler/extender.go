package scheduler

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/log"
)

// Extender reperesents a scheduler extender.
type Extender struct {
	// Name identifies the Extender.
	Name string

	// Filter filters out the nodes that cannot run the given pod.
	// This function can be nil.
	Filter func(api.ExtenderArgs) api.ExtenderFilterResult

	// Prioritize ranks each node that has passes the filtering stage.
	// The weighted scores are summed up and the total score is used for the node selection.
	Prioritize func(api.ExtenderArgs) api.HostPriorityList
	Weight     int

	// NodeCacheCapable specifies that the extender is capable of caching node information, so the
	// scheduler should only send minimal information about the eligible nodes assuming that the
	// extender already cached full details of all nodes in the cluster.
	// Specifically, ExtenderArgs.NodeNames is populated only if NodeCacheCapable == true, and
	// ExtenderArgs.Nodes.Items is populated only if NodeCacheCapable == false.
	NodeCacheCapable bool

	// Ignorable specifies if the extender is ignorable, i.e., scheduling should not fail when the
	// extender returns an error.
	Ignorable bool
}

func (ext *Extender) filter(
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	failedPredicateMap core.FailedPredicateMap) ([]*v1.Node, error) {

	if ext.Filter == nil {
		return nodes, nil
	}

	log.L.Tracef("Extender %q: Filtering nodes %v", ext.Name, nodes)

	args := buildExtenderArgs(pod, nodes, ext.NodeCacheCapable)
	// FIXME: args.NodeNames may be not populated
	log.L.Debugf("Extender %q: Filtering nodes %v", ext.Name, args.NodeNames)

	result := ext.Filter(args)

	nodes = make([]*v1.Node, 0, len(nodes))
	if ext.NodeCacheCapable {
		for _, name := range *result.NodeNames {
			nodes = append(nodes, nodeInfoMap[name].Node())
		}
	} else {
		for _, node := range result.Nodes.Items {
			nodes = append(nodes, &node)
		}
	}

	for failedNodeName, failedMsg := range result.FailedNodes {
		if _, found := failedPredicateMap[failedNodeName]; !found {
			failedPredicateMap[failedNodeName] = []predicates.PredicateFailureReason{}
		}
		failedPredicateMap[failedNodeName] = append(failedPredicateMap[failedNodeName], predicates.NewFailureReason(failedMsg))
	}

	if result.Error != "" {
		if ext.Ignorable {
			log.L.Warnf("Skipping ext %q as it returned error %q and has ignorable flag set", ext.Name, result.Error)
		} else {
			return nodes, errors.New(result.Error)
		}
	}

	log.L.Tracef("Filtered nodes %v", nodes)
	log.L.Debugf("Filtered nodes %v", result.NodeNames)

	return nodes, nil
}

func (ext *Extender) prioritize(pod *v1.Pod, nodes []*v1.Node, prioMap map[string]int) {
	if ext.Prioritize == nil {
		return
	}

	log.L.Tracef("Extender %q: Prioritizing nodes %v", ext.Name, nodes)

	args := buildExtenderArgs(pod, nodes, ext.NodeCacheCapable)
	// FIXME: args.NodeNames may be not populated
	log.L.Debugf("Extender %q: Prioritizing nodes %v", ext.Name, args.NodeNames)

	result := ext.Prioritize(args)

	log.L.Debugf("Prioritized %v", result)
	for _, prio := range result {
		prioMap[prio.Host] += prio.Score * ext.Weight
	}
}

func buildExtenderArgs(pod *v1.Pod, nodes []*v1.Node, nodeCacheCapable bool) api.ExtenderArgs {
	nodeList := v1.NodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodeList",
			APIVersion: "v1",
		},
		// ListMeta: metav1.ListMeta{},
		Items: make([]v1.Node, 0, len(nodes)),
	}
	nodeNames := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if nodeCacheCapable {
			nodeNames = append(nodeNames, node.Name)
		} else {
			nodeList.Items = append(nodeList.Items, *node)
		}
	}

	return api.ExtenderArgs{
		Pod:       pod,
		Nodes:     &nodeList,
		NodeNames: &nodeNames,
	}
}
