package scheduler

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/log"
)

// type Scheduler interface {
// 	Schedule(podLister algorithm.PodLister, nodeLister algorithm.NodeLister, nodeInfoMap map[string]*nodeinfo.NodeInfo) (core.ScheduleResult, error)
// }

// Scheduler makes scheduling decision for each given pod.
//
// It mimics "k8s.io/pkg/Scheduler/Scheduler/core".genericScheduler, which implements
// "k8s.io/pkg/Scheduler/Scheduler/core".ScheduleAlgorithm.
type Scheduler struct {
	extenders []Extender

	predicates   map[string]predicates.FitPredicate
	prioritizers []priorities.PriorityConfig

	lastNodeIndex uint64
}

// NewScheduler creates a new Scheduler.
func NewScheduler() Scheduler {
	return Scheduler{
		predicates: map[string]predicates.FitPredicate{},
	}
}

// AddExtender adds an extender to this Scheduler.
func (sched *Scheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// AddPredicate adds a predicate plugin to this Scheduler.
func (sched *Scheduler) AddPredicate(name string, predicate predicates.FitPredicate) {
	sched.predicates[name] = predicate
}

// AddPrioritizer adds a prioritizer plugin to this Scheduler.
func (sched *Scheduler) AddPrioritizer(prioritizer priorities.PriorityConfig) {
	sched.prioritizers = append(sched.prioritizers, prioritizer)
}

// Schedule makes scheduling decision for the given pod and nodes.
func (sched *Scheduler) Schedule(
	pod *v1.Pod,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) (core.ScheduleResult, error) {

	result := core.ScheduleResult{}
	nodes, err := nodeLister.List()
	if err != nil {
		return result, err
	}
	if len(nodes) == 0 {
		return result, core.ErrNoNodesAvailable
	}

	nodesFiltered, failedPredicateMap, err := sched.filter(pod, nodes, nodeInfoMap)
	if err != nil {
		return result, err
	}

	switch len(nodesFiltered) {
	case 0:
		return result, &core.FitError{
			Pod:              pod,
			NumAllNodes:      len(nodes),
			FailedPredicates: failedPredicateMap,
		}
	case 1:
		return core.ScheduleResult{
			SuggestedHost:  nodesFiltered[0].Name,
			EvaluatedNodes: 1 + len(failedPredicateMap),
			FeasibleNodes:  1,
		}, nil
	}

	prios, err := sched.prioritize(pod, nodesFiltered, nodeInfoMap)
	if err != nil {
		return result, err
	}
	host, err := sched.selectHost(prios)

	return core.ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(nodesFiltered) + len(failedPredicateMap),
		FeasibleNodes:  len(nodesFiltered),
	}, err
}

// func (sched *Scheduler) Preempt(
// 	pod *v1.Pod,
// 	nodeLister algorithm.NodeLister,
// 	err error) (selectedNode *v1.Node,
// 	preemptedPods []*v1.Pod,
// 	cleanupNominatedPods []*v1.Pod, e error) {
// }

func (sched *Scheduler) filter(
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]*v1.Node, core.FailedPredicateMap, error) {

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Filtering nodes %v", nodeNames)

	failedPredicateMap := core.FailedPredicateMap{}
	filteredNodes := nodes

	// In-process plugins
	errs := errors.MessageCountMap{}
	for name, p := range sched.predicates {
		filteredNodes = callPredicatePlugin(name, &p, pod, filteredNodes, nodeInfoMap, failedPredicateMap, errs)
		if len(filteredNodes) == 0 {
			break
		}
	}

	if len(errs) > 0 {
		return []*v1.Node{}, core.FailedPredicateMap{}, errors.CreateAggregateFromMessageCountMap(errs)
	}

	// Extenders
	if len(filteredNodes) > 0 && len(sched.extenders) > 0 {
		for _, extender := range sched.extenders {
			var err error
			filteredNodes, err = extender.filter(pod, filteredNodes, nodeInfoMap, failedPredicateMap)
			if err != nil {
				return []*v1.Node{}, core.FailedPredicateMap{}, err
			}

			if len(filteredNodes) == 0 {
				break
			}
		}
	}

	nodeNames = make([]string, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Filtered %v", nodeNames)

	return filteredNodes, failedPredicateMap, nil
}

func (sched *Scheduler) prioritize(
	pod *v1.Pod,
	filteredNodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) (api.HostPriorityList, error) {

	// FIXME: Make nodeNames only when debug logging is enabled.
	nodeNames := make([]string, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	log.L.Debugf("Prioritizing nodes %v", nodeNames)

	prioList := make(api.HostPriorityList, len(filteredNodes), len(filteredNodes))

	// If no priority configs are provided, then the EqualPriority function is applied
	// This is required to generate the priority list in the required format
	if len(sched.prioritizers) == 0 && len(sched.extenders) == 0 {
		for i, node := range filteredNodes {
			prio, err := core.EqualPriorityMap(pod, &dummyPriorityMetadata{}, nodeInfoMap[node.Name])
			if err != nil {
				return api.HostPriorityList{}, err
			}
			prioList[i] = prio
		}
		return prioList, nil
	}

	for i, node := range filteredNodes {
		prioList[i] = api.HostPriority{Host: node.Name, Score: 0}
	}

	errs := []error{}

	// In-process plugins
	for _, prioritizer := range sched.prioritizers {
		prios := callPrioritizePlugin(&prioritizer, pod, filteredNodes, nodeInfoMap, errs)
		for i, prio := range prios {
			prioList[i].Score += prio.Score
		}
	}

	if len(errs) > 0 {
		return api.HostPriorityList{}, errors.NewAggregate(errs)
	}

	// Extenders
	if len(sched.extenders) > 0 {
		prioMap := map[string]int{}
		for _, extender := range sched.extenders {
			extender.prioritize(pod, filteredNodes, prioMap)
		}

		for i, prio := range prioList {
			prioList[i].Score += prioMap[prio.Host]
		}
	}

	log.L.Debugf("Prioritized %v", prioList)

	return prioList, nil
}

func (sched *Scheduler) selectHost(priorities api.HostPriorityList) (string, error) {
	if len(priorities) == 0 {
		return "", fmt.Errorf("empty priorities")
	}

	maxScores := findMaxScores(priorities)
	idx := int(sched.lastNodeIndex % uint64(len(maxScores)))
	sched.lastNodeIndex++

	return priorities[maxScores[idx]].Host, nil
}

func findMaxScores(priorities api.HostPriorityList) []int {
	maxScoreIndexes := make([]int, 0, len(priorities)/2)
	maxScore := priorities[0].Score

	for idx, prio := range priorities {
		if prio.Score > maxScore {
			maxScore = prio.Score
			maxScoreIndexes = maxScoreIndexes[:0]
			maxScoreIndexes = append(maxScoreIndexes, idx)
		} else if prio.Score == maxScore {
			maxScoreIndexes = append(maxScoreIndexes, idx)
		}
	}

	return maxScoreIndexes
}
