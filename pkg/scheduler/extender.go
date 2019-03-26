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
	"errors"
	"fmt"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	l "github.com/pfnet-research/k8s-cluster-simulator/pkg/log"
)

// Extender reperesents a scheduler extender.
type Extender struct {
	// Name identifies this Extender.
	Name string

	// Filter filters out the nodes that cannot run the given pod in api.ExtenderArgs.
	// This function can be nil.
	Filter func(api.ExtenderArgs) api.ExtenderFilterResult

	// Prioritize ranks each node that has passes the filtering stage.
	// The weighted scores are summed up and the total score is used for the node selection.
	Prioritize func(api.ExtenderArgs) api.HostPriorityList
	Weight     int

	// NodeCacheCapable specifies that this Extender is capable of caching node information, so the
	// scheduler should only send minimal information about the eligible nodes assuming that the
	// extender already cached full details of all nodes in the cluster.
	// Specifically, ExtenderArgs.NodeNames is populated iff NodeCacheCapable == true, and
	// ExtenderArgs.Nodes.Items is populated iff NodeCacheCapable == false.
	NodeCacheCapable bool

	// Ignorable specifies whether the extender is ignorable (i.e. the scheduler process should not
	// fail when this extender returns an error).
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

	log.L.Tracef("Extender %s: Filtering nodes %v", ext.Name, nodes)

	// Build an argument and call this extender.
	args := buildExtenderArgs(pod, nodes, ext.NodeCacheCapable)

	if l.IsDebugEnabled() {
		nodeNames := make([]string, 0, len(nodes))
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		log.L.Debugf("Extender %s: Filtering nodes %v", ext.Name, nodeNames)
	}

	result := ext.Filter(args)

	// Arrange the returned values.
	nodes = make([]*v1.Node, 0, len(nodes))
	if ext.NodeCacheCapable {
		for _, name := range *result.NodeNames {
			nodeInfo, ok := nodeInfoMap[name]
			if !ok {
				return []*v1.Node{}, fmt.Errorf("No node named %s", name)
			}
			nodes = append(nodes, nodeInfo.Node())
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
			log.L.Warnf("Skipping extender %q as it returned error %q and has ignorable flag set", ext.Name, result.Error)
		} else {
			return nodes, errors.New(result.Error)
		}
	}

	log.L.Tracef("Extender %s: Filtered nodes %v", ext.Name, nodes)
	if l.IsDebugEnabled() {
		nodeNames := make([]string, 0, len(nodes))
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		log.L.Debugf("Extender %s: Filtered nodes %v", ext.Name, nodeNames)
	}

	return nodes, nil
}

func (ext *Extender) prioritize(pod *v1.Pod, nodes []*v1.Node, prioMap map[string]int) {
	if ext.Prioritize == nil {
		return
	}

	log.L.Tracef("Extender %s: Prioritizing nodes %v", ext.Name, nodes)

	// Build an argument and call this extender.
	args := buildExtenderArgs(pod, nodes, ext.NodeCacheCapable)

	if l.IsDebugEnabled() {
		nodeNames := make([]string, 0, len(nodes))
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		log.L.Debugf("Extender %s: Prioritizing nodes %v", ext.Name, nodeNames)
	}

	result := ext.Prioritize(args)

	// Sum up the returned values.
	log.L.Debugf("Extender %s: Prioritized %v", ext.Name, result)
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
