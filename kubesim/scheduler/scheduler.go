package scheduler

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"

	"github.com/ordovicia/kubernetes-simulator/log"
)

// Extender reperesents a scheduler extender.
type Extender struct {
	Name string

	// Filter filters out the nodes that cannot run the given pod.
	// This function can be nil.
	Filter func(api.ExtenderArgs) api.ExtenderFilterResult

	// Prioritize ranks each node that has passes the filtering stage.
	// The weighted scores are summed up and the total score is used for the node selection.
	Prioritize func(api.ExtenderArgs) api.HostPriorityList
	Weight     int

	NodeCacheCapable bool
}

// Scheduler makes scheduling decision for each given pod.
//
// It mimics "k8s.io/pkg/Scheduler/Scheduler/core".genericScheduler, which implements
// "k8s.io/pkg/Scheduler/Scheduler/core".ScheduleAlgorithm
type Scheduler struct {
	nodes         map[string]*v1.Node
	extenders     []Extender
	lastNodeIndex uint64
}

// NewScheduler creates new Scheduler with the nodes.
func NewScheduler(nodes map[string]*v1.Node) Scheduler {
	return Scheduler{
		nodes:         nodes,
		extenders:     []Extender{},
		lastNodeIndex: 0,
	}
}

// AddExtender adds an extender to this Scheduler.
func (sched *Scheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// Schedule makes scheduling decision for the given pod and nodes.
func (sched *Scheduler) Schedule(pod *v1.Pod, nodeLister algorithm.NodeLister) (core.ScheduleResult, error) {
	log.L.Tracef("Trying to schedule pod %v", pod)

	result := core.ScheduleResult{}

	nodes, err := nodeLister.List()
	if err != nil {
		return result, err
	}
	if len(nodes) == 0 {
		return result, core.ErrNoNodesAvailable
	}

	nodesFiltered, failedPredicateMap, err := sched.filter(pod, nodes)
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

	priorities := sched.prioritize(pod, nodesFiltered)
	host, err := sched.selectHost(priorities)

	return core.ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(nodesFiltered) + len(failedPredicateMap),
		FeasibleNodes:  len(nodesFiltered),
	}, err
}

// func (sched *Scheduler) Preempt(pod *v1.Pod, nodeLister algorithm.NodeLister, err error) (selectedNode *v1.Node, preemptedPods []*v1.Pod, cleanupNominatedPods []*v1.Pod, err error)

func (sched *Scheduler) filter(pod *v1.Pod, nodes []*v1.Node) ([]*v1.Node, core.FailedPredicateMap, error) {
	failedPredicateMap := core.FailedPredicateMap{}

	for _, extender := range sched.extenders {
		if extender.Filter == nil {
			continue
		}

		log.L.Tracef("Extender %q: Filtering nodes %v", extender.Name, nodes)

		args := buildExtenderArgs(pod, nodes, extender.NodeCacheCapable)
		result := extender.Filter(args)

		nodes = []*v1.Node{}
		if extender.NodeCacheCapable {
			for _, name := range *result.NodeNames {
				nodes = append(nodes, sched.nodes[name])
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
			return nodes, failedPredicateMap, errors.New(result.Error)
		}

		log.L.Tracef("Filtered nodes %v", nodes)

		if len(nodes) == 0 {
			break
		}
	}

	return nodes, failedPredicateMap, nil
}

func (sched *Scheduler) prioritize(pod *v1.Pod, nodes []*v1.Node) api.HostPriorityList {
	prioMap := map[string]int{}

	for _, extender := range sched.extenders {
		if extender.Prioritize == nil {
			continue
		}

		log.L.Tracef("Extender %q: Prioritizing nodes %v", extender.Name, nodes)

		args := buildExtenderArgs(pod, nodes, extender.NodeCacheCapable)
		result := extender.Prioritize(args)

		log.L.Tracef("Prioritized %v", result)
		for _, prio := range result {
			prioMap[prio.Host] += prio.Score * extender.Weight
		}
	}

	prioList := api.HostPriorityList{}
	for name, score := range prioMap {
		prioList = append(prioList, api.HostPriority{Host: name, Score: score})
	}

	return prioList
}

func buildExtenderArgs(pod *v1.Pod, nodes []*v1.Node, nodeCacheCapable bool) api.ExtenderArgs {
	nodeList := v1.NodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodeList",
			APIVersion: "v1",
		},
		// ListMeta: metav1.ListMeta{},
		Items: []v1.Node{},
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
