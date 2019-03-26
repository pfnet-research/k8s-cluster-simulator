// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
)

// dummyPredicateMetadata implements predicates.PredicateMetadata interface.
type dummyPredicateMetadata struct{}
type dummyPriorityMetadata struct{}

func (d *dummyPredicateMetadata) ShallowCopy() predicates.PredicateMetadata             { return d }
func (d *dummyPredicateMetadata) AddPod(pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo) error { return nil }
func (d *dummyPredicateMetadata) RemovePod(pod *v1.Pod) error                           { return nil }

const workerNum = 16

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

	nodesNum := int32(len(nodes))

	filtered := make([]*v1.Node, nodesNum)
	filteredLen := int32(0)

	errs := errors.MessageCountMap{}
	var predicateResultLock sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())

	// Run predicate plugins in parallel along nodes.
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
		errs        []error
		errsLock    = sync.Mutex{}
		appendError = func(err error) {
			errsLock.Lock()
			defer errsLock.Unlock()
			errs = append(errs, err)
		}
	)

	resultList := make([]api.HostPriorityList, len(prioritizers))
	for i := range prioritizers {
		resultList[i] = make(api.HostPriorityList, len(nodes))
	}

	// Run map phases of prioritizer plugins in parallel along nodes.
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

	// Run reduce phases of prioritizer plugins in parallel along plugins.
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

	// Sum up all scores along nodes.
	prioList = make(api.HostPriorityList, 0, len(nodes))
	for nodeIdx := range nodes {
		prioList = append(prioList, api.HostPriority{Host: nodes[nodeIdx].Name, Score: 0})
		for prioIdx := range prioritizers {
			prioList[nodeIdx].Score += resultList[prioIdx][nodeIdx].Score * prioritizers[prioIdx].Weight
		}
	}

	return prioList, nil
}
