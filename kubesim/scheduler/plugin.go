package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/k8s-cluster-simulator/kubesim/queue"
)

// dummyPredicateMetadata implements predicates.PredicateMetadata.
type dummyPredicateMetadata struct{}
type dummyPriorityMetadata struct{}

func (d *dummyPredicateMetadata) ShallowCopy() predicates.PredicateMetadata             { return d }
func (d *dummyPredicateMetadata) AddPod(pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo) error { return nil }
func (d *dummyPredicateMetadata) RemovePod(pod *v1.Pod) error                           { return nil }

/*
func callPredicatePlugin(
	name string,
	pred predicates.FitPredicate,
	pod *v1.Pod,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	failedPredicateMap core.FailedPredicateMap,
	errs errors.MessageCountMap) (filteredNodes []*v1.Node, err error) {

	log.L.Tracef("Plugin %s: Predicating nodes %v", name, nodes)

	if log.IsDebugEnabled() {
		nodeNames := make([]string, 0, len(nodes))
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		log.L.Debugf("Plugin %s: Predicating nodes %v", name, nodeNames)
	}

	filteredNodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeInfo, ok := nodeInfoMap[node.Name]
		if !ok {
			return []*v1.Node{}, fmt.Errorf("No node named %s", node.Name)
		}

		fits, failureReason, err := pred(pod, &dummyPredicateMetadata{}, nodeInfo)
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

	return filteredNodes, nil
}

func callPrioritizePlugin(
	prioritizer *priorities.PriorityConfig,
	pod *v1.Pod,
	filteredNodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	errs []error) (api.HostPriorityList, error) {

	log.L.Tracef("Plugin %s: Prioritizing nodes %v", prioritizer.Name, filteredNodes)

	if log.IsDebugEnabled() {
		nodeNames := make([]string, 0, len(filteredNodes))
		for _, node := range filteredNodes {
			nodeNames = append(nodeNames, node.Name)
		}
		log.L.Debugf("Plugin %s: Prioritizing nodes %v", prioritizer.Name, nodeNames)
	}

	prios := make(api.HostPriorityList, 0, len(filteredNodes))
	for _, node := range filteredNodes {
		nodeInfo, ok := nodeInfoMap[node.Name]
		if !ok {
			return api.HostPriorityList{}, fmt.Errorf("No node named %s", node.Name)
		}

		prio, err := prioritizer.Map(pod, &dummyPriorityMetadata{}, nodeInfo)
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

	return prios, nil
}
*/

const workerNum = 8

func filterWithPlugins(
	pod *v1.Pod,
	preds map[string]predicates.FitPredicate,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	podQueue queue.PodQueue,
) ([]*v1.Node, core.FailedPredicateMap, error) {
	failedPredicateMap := core.FailedPredicateMap{}

	if len(preds) == 0 {
		return nodes, failedPredicateMap, nil
	}

	nodesNum := int32(len(nodes)) // TODO: g.numFeasibleNodesToFind(allNodes)

	filtered := make([]*v1.Node, nodesNum)
	filteredLen := int32(0)

	errs := errors.MessageCountMap{}
	var predicateResultLock sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())

	workqueue.ParallelizeUntil(ctx, workerNum, int(nodesNum), func(i int) {
		nodeName := nodes[i].Name
		nodeInfo, ok := nodeInfoMap[nodeName]
		if !ok {
			err := fmt.Errorf("No node named %q", nodeName)
			predicateResultLock.Lock()
			defer predicateResultLock.Unlock()
			errs[err.Error()]++
			return
		}

		fits, failedPredicates, err := podFitsOnNode(
			pod,
			preds,
			nodeInfo,
			podQueue,
		)
		if err != nil {
			predicateResultLock.Lock()
			defer predicateResultLock.Unlock()
			errs[err.Error()]++
			return
		}

		if fits {
			length := atomic.AddInt32(&filteredLen, 1)
			if length > nodesNum {
				cancel()
				atomic.AddInt32(&filteredLen, -1)
			} else {
				filtered[length-1] = nodes[i]
			}
		} else {
			predicateResultLock.Lock()
			defer predicateResultLock.Unlock()
			failedPredicateMap[nodeName] = failedPredicates
		}
	})

	if len(errs) > 0 {
		return []*v1.Node{}, core.FailedPredicateMap{}, errors.CreateAggregateFromMessageCountMap(errs)
	}

	return filtered[:filteredLen], failedPredicateMap, nil
}

func prioritizeWithPlugins(
	pod *v1.Pod,
	prioritizers []priorities.PriorityConfig,
	nodes []*v1.Node,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	podQueue queue.PodQueue,
) (prioList api.HostPriorityList, err error) {
	var (
		errs     []error
		errsLock = sync.Mutex{}
	)

	appendError := func(err error) {
		errsLock.Lock()
		defer errsLock.Unlock()
		errs = append(errs, err)
	}

	resultList := make([]api.HostPriorityList, len(prioritizers))
	for i := range prioritizers {
		resultList[i] = make(api.HostPriorityList, len(nodes))
	}

	// Map
	workqueue.ParallelizeUntil(context.TODO(), workerNum, len(nodes), func(nodeIdx int) {
		nodeName := nodes[nodeIdx].Name
		nodeInfo, ok := nodeInfoMap[nodeName]
		if !ok {
			appendError(fmt.Errorf("No node named %q", nodes[nodeIdx].Name))
			return
		}

		for prioIdx := range prioritizers {
			if prioritizers[prioIdx].Function != nil {
				continue
			}

			var err error
			resultList[prioIdx][nodeIdx], err = prioritizers[prioIdx].Map(pod, &dummyPriorityMetadata{}, nodeInfo)
			if err != nil {
				appendError(err)
				resultList[prioIdx][nodeIdx].Host = nodeName
			}
		}
	})

	// Reduce
	wg := sync.WaitGroup{}
	for prioIdx := range prioritizers {
		if prioritizers[prioIdx].Reduce == nil {
			continue
		}

		wg.Add(1)
		go func(prioIdx int) {
			defer wg.Done()
			if err := prioritizers[prioIdx].Reduce(pod, &dummyPriorityMetadata{}, nodeInfoMap, resultList[prioIdx]); err != nil {
				appendError(err)
			}
		}(prioIdx)
	}

	wg.Wait()
	if len(errs) != 0 {
		return api.HostPriorityList{}, errors.NewAggregate(errs)
	}

	// Summarize all scores.
	prioList = make(api.HostPriorityList, 0, len(nodes))

	for nodeIdx := range nodes {
		prioList = append(prioList, api.HostPriority{Host: nodes[nodeIdx].Name, Score: 0})
		for prioIdx := range prioritizers {
			prioList[nodeIdx].Score += resultList[prioIdx][nodeIdx].Score * prioritizers[prioIdx].Weight
		}
	}

	return prioList, nil
}
