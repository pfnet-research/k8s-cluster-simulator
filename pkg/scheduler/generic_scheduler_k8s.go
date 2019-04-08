/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Modifications copyright 2019 Preferred Networks, Inc.
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

// All functions in this file were copied from
// k8s.io/kubernetes/pkg/scheduler/core/generic_scheduler.go by the authors of
// k8s-cluster-simulator, and modified so that they would be compatible with k8s-cluster-simulator.

package scheduler

import (
	"errors"
	"math"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	kutil "k8s.io/kubernetes/pkg/scheduler/util"

	l "github.com/pfnet-research/k8s-cluster-simulator/pkg/log"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

func (sched *GenericScheduler) selectHost(priorities api.HostPriorityList) (string, error) {
	if len(priorities) == 0 {
		return "", errors.New("Empty priorities")
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

func podEligibleToPreemptOthers(preemptor *v1.Pod, nodeInfoMap map[string]*nodeinfo.NodeInfo) bool {
	nomNodeName := preemptor.Status.NominatedNodeName
	if len(nomNodeName) > 0 {
		if nodeInfo, ok := nodeInfoMap[nomNodeName]; ok {
			for _, p := range nodeInfo.Pods() {
				if p.DeletionTimestamp != nil && util.PodPriority(p) < util.PodPriority(preemptor) {
					// There is a terminating pod on the nominated node.
					return false
				}
			}
		}
	}

	return true
}

func nodesWherePreemptionMightHelp(nodes []*v1.Node, failedPredicatesMap core.FailedPredicateMap) []*v1.Node {
	potentialNodes := []*v1.Node{}

	for _, node := range nodes {
		unresolvableReasonExist := false
		failedPredicates, _ := failedPredicatesMap[node.Name]

		for _, failedPredicate := range failedPredicates {
			switch failedPredicate {
			case
				predicates.ErrNodeSelectorNotMatch,
				predicates.ErrPodAffinityRulesNotMatch,
				predicates.ErrPodNotMatchHostName,
				predicates.ErrTaintsTolerationsNotMatch,
				predicates.ErrNodeLabelPresenceViolated,
				// Node conditions won't change when scheduler simulates removal of preemption victims.
				// So, it is pointless to try nodes that have not been able to host the pod due to node
				// conditions. These include ErrNodeNotReady, ErrNodeUnderPIDPressure, ErrNodeUnderMemoryPressure, ....
				predicates.ErrNodeNotReady,
				predicates.ErrNodeNetworkUnavailable,
				predicates.ErrNodeUnderDiskPressure,
				predicates.ErrNodeUnderPIDPressure,
				predicates.ErrNodeUnderMemoryPressure,
				predicates.ErrNodeUnschedulable,
				predicates.ErrNodeUnknownCondition,
				predicates.ErrVolumeZoneConflict,
				predicates.ErrVolumeNodeConflict,
				predicates.ErrVolumeBindConflict:
				unresolvableReasonExist = true
				break
			}
		}

		if !unresolvableReasonExist {
			log.L.Tracef("Node %v is a potential node for preemption.", node)
			log.L.Debugf("Node %s is a potential node for preemption.", node.Name)
			potentialNodes = append(potentialNodes, node)
		}
	}

	return potentialNodes
}

func (sched *GenericScheduler) selectNodesForPreemption(
	preemptor *v1.Pod,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	potentialNodes []*v1.Node,
	podQueue queue.PodQueue,
	// pdbs []*policy.PodDisruptionBudget,
) (map[*v1.Node]*api.Victims, error) {
	nodeToVictims := map[*v1.Node]*api.Victims{}

	for _, node := range potentialNodes {
		pods, numPDBViolations, fits := sched.selectVictimsOnNode(preemptor, nodeInfoMap[node.Name], podQueue /* , pdbs */)
		if fits {
			nodeToVictims[node] = &api.Victims{
				Pods:             pods,
				NumPDBViolations: numPDBViolations,
			}
		}
	}

	return nodeToVictims, nil
}

func (sched *GenericScheduler) selectVictimsOnNode(
	preemptor *v1.Pod,
	nodeInfo *nodeinfo.NodeInfo,
	podQueue queue.PodQueue,
	// pdbs []*policy.PodDisruptionBudget,
) (pods []*v1.Pod, numPDBViolations int, fits bool) {
	if nodeInfo == nil {
		return nil, 0, false
	}

	potentialVictims := kutil.SortableList{CompFunc: kutil.HigherPriorityPod}
	nodeInfoCopy := nodeInfo.Clone()

	removePod := func(p *v1.Pod) {
		nodeInfoCopy.RemovePod(p)
	}

	addPod := func(p *v1.Pod) {
		nodeInfoCopy.AddPod(p)
	}

	podPriority := util.PodPriority(preemptor)
	for _, p := range nodeInfoCopy.Pods() {
		if util.PodPriority(p) < podPriority {
			potentialVictims.Items = append(potentialVictims.Items, p)
			removePod(p)
		}
	}
	potentialVictims.Sort()

	if fits, _, err := podFitsOnNode(preemptor, sched.predicates, nodeInfoCopy, podQueue); !fits {
		if err != nil {
			log.L.Warnf("Encountered error while selecting victims on node %s: %v", nodeInfoCopy.Node().Name, err)
		}

		log.L.Debugf(
			"Preemptor does not fit in node %s even if all lower-priority pods were removed",
			nodeInfoCopy.Node().Name)
		return nil, 0, false
	}

	var victims []*v1.Pod
	// numViolatingVictim := 0

	// // Try to reprieve as many pods as possible. We first try to reprieve the PDB
	// // violating victims and then other non-violating ones. In both cases, we start
	// // from the highest priority victims.
	// violatingVictims, nonViolatingVictims := filterPodsWithPDBViolation(potentialVictims.Items, pdbs)

	reprievePod := func(p *v1.Pod) bool {
		addPod(p)
		fits, _, _ := podFitsOnNode(preemptor, sched.predicates, nodeInfoCopy, podQueue)
		if !fits {
			removePod(p)
			victims = append(victims, p)

			if l.IsDebugEnabled() {
				key, err := util.PodKey(p)
				if err != nil {
					log.L.Warnf("Encountered error while building key of pod %v: %v", p, err)
					return fits
				}
				log.L.Debugf("Pod %s is a potential preemption victim on node %s.", key, nodeInfoCopy.Node().Name)
			}
		}

		return fits
	}

	for _, p := range /* violatingVictims */ potentialVictims.Items {
		if !reprievePod(p.(*v1.Pod)) {
			// numViolatingVictim++
		}
	}

	// // Now we try to reprieve non-violating victims.
	// for _, p := range nonViolatingVictims {
	// 	reprievePod(p)
	// }

	return victims /* numViolatingVictim */, 0, true
}

func podFitsOnNode(
	pod *v1.Pod,
	preds map[string]predicates.FitPredicate,
	nodeInfo *nodeinfo.NodeInfo,
	podQueue queue.PodQueue,
) (bool, []predicates.PredicateFailureReason, error) {
	var failedPredicates []predicates.PredicateFailureReason

	podsAdded := false
	for i := 0; i < 2; i++ {
		nodeInfoToUse := nodeInfo
		if i == 0 {
			podsAdded, nodeInfoToUse = addNominatedPods(pod, nodeInfo, podQueue)
		} else if !podsAdded || len(failedPredicates) != 0 {
			break
		}

		for _, pred := range preds {
			fit, reasons, err := pred(pod, &dummyPredicateMetadata{}, nodeInfoToUse)

			if err != nil {
				return false, []predicates.PredicateFailureReason{}, err
			}

			if !fit {
				failedPredicates = append(failedPredicates, reasons...)
				break
			}
		}
	}

	return len(failedPredicates) == 0, failedPredicates, nil
}

func addNominatedPods(
	pod *v1.Pod, nodeInfo *nodeinfo.NodeInfo, podQueue queue.PodQueue,
) (bool, *nodeinfo.NodeInfo) {
	nominatedPods := podQueue.NominatedPods(nodeInfo.Node().Name)
	if len(nominatedPods) == 0 {
		return false, nodeInfo
	}

	nodeInfoOut := nodeInfo.Clone()
	for _, p := range nominatedPods {
		if util.PodPriority(p) >= util.PodPriority(pod) && p.UID != pod.UID {
			nodeInfoOut.AddPod(p)
		}
	}

	return true, nodeInfoOut
}

func pickOneNodeForPreemption(nodesToVictims map[*v1.Node]*api.Victims) *v1.Node {
	if len(nodesToVictims) == 0 {
		return nil
	}

	// minNumPDBViolatingPods := math.MaxInt32
	var minNodes1 []*v1.Node
	lenNodes1 := 0
	for node := range nodesToVictims {
		// if len(victims.Pods) == 0 {
		// 	// We found a node that doesn't need any preemption. Return it!
		// 	// This should happen rarely when one or more pods are terminated between
		// 	// the time that scheduler tries to schedule the pod and the time that
		// 	// preemption logic tries to find nodes for preemption.
		// 	return node
		// }

		// numPDBViolatingPods := victims.NumPDBViolations
		// if numPDBViolatingPods < minNumPDBViolatingPods {
		// 	minNumPDBViolatingPods = numPDBViolatingPods
		// 	minNodes1 = nil
		// 	lenNodes1 = 0
		// }
		// if numPDBViolatingPods == minNumPDBViolatingPods {
		// 	minNodes1 = append(minNodes1, node)
		// 	lenNodes1++
		// }

		minNodes1 = append(minNodes1, node)
		lenNodes1++
	}
	if lenNodes1 == 1 {
		return minNodes1[0]
	}

	// There are more than one node with minimum number PDB violating pods. Find
	// the one with minimum highest priority victim.
	minHighestPriority := int32(math.MaxInt32)
	var minNodes2 = make([]*v1.Node, lenNodes1)
	lenNodes2 := 0
	for i := 0; i < lenNodes1; i++ {
		node := minNodes1[i]
		victims := nodesToVictims[node]
		// highestPodPriority is the highest priority among the victims on this node.
		highestPodPriority := util.PodPriority(victims.Pods[0])
		if highestPodPriority < minHighestPriority {
			minHighestPriority = highestPodPriority
			lenNodes2 = 0
		}
		if highestPodPriority == minHighestPriority {
			minNodes2[lenNodes2] = node
			lenNodes2++
		}
	}
	if lenNodes2 == 1 {
		return minNodes2[0]
	}

	// There are a few nodes with minimum highest priority victim. Find the
	// smallest sum of priorities.
	minSumPriorities := int64(math.MaxInt64)
	lenNodes1 = 0
	for i := 0; i < lenNodes2; i++ {
		var sumPriorities int64
		node := minNodes2[i]
		for _, pod := range nodesToVictims[node].Pods {
			// We add MaxInt32+1 to all priorities to make all of them >= 0. This is
			// needed so that a node with a few pods with negative priority is not
			// picked over a node with a smaller number of pods with the same negative
			// priority (and similar scenarios).
			sumPriorities += int64(util.PodPriority(pod)) + int64(math.MaxInt32+1)
		}
		if sumPriorities < minSumPriorities {
			minSumPriorities = sumPriorities
			lenNodes1 = 0
		}
		if sumPriorities == minSumPriorities {
			minNodes1[lenNodes1] = node
			lenNodes1++
		}
	}
	if lenNodes1 == 1 {
		return minNodes1[0]
	}

	// There are a few nodes with minimum highest priority victim and sum of priorities.
	// Find one with the minimum number of pods.
	minNumPods := math.MaxInt32
	lenNodes2 = 0
	for i := 0; i < lenNodes1; i++ {
		node := minNodes1[i]
		numPods := len(nodesToVictims[node].Pods)
		if numPods < minNumPods {
			minNumPods = numPods
			lenNodes2 = 0
		}
		if numPods == minNumPods {
			minNodes2[lenNodes2] = node
			lenNodes2++
		}
	}
	// At this point, even if there are more than one node with the same score,
	// return the first one.
	if lenNodes2 > 0 {
		return minNodes2[0]
	}

	log.L.Error("Error in logic of node scoring for preemption. We should never reach here!")
	return nil
}

func getLowerPriorityNominatedPods(pod *v1.Pod, nodeName string, podQueue queue.PodQueue) []*v1.Pod {
	pods := podQueue.NominatedPods(nodeName)
	if len(pods) == 0 {
		return nil
	}

	var lowerPriorityPods []*v1.Pod
	podPriority := util.PodPriority(pod)
	for _, p := range pods {
		if util.PodPriority(p) < podPriority {
			lowerPriorityPods = append(lowerPriorityPods, p)
		}
	}
	return lowerPriorityPods
}
