package kubesim

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/ordovicia/kubernetes-simulator/log"
)

type Scheduler struct {
	nodeInfoMap map[string]*nodeinfo.NodeInfo
	extenders   []algorithm.SchedulerExtender
}

func (sched *Scheduler) Schedule(pod *v1.Pod, nodeLister algorithm.NodeLister) (result core.ScheduleResult, err error) {
	log.L.Tracef("Trying to schedule pod %v", pod)

	nodes, err := nodeLister.List()
	if err != nil {
		return result, err
	}
	if len(nodes) == 0 {
		return result, core.ErrNoNodesAvailable
	}

	// result := core.ScheduleResult{
	// 	SuggestedHost:  "foo",
	// 	EvaluatedNodes: 0,
	// 	FeasibleNodes:  0,
	// }

	return result, nil
}

// func (sched *Scheduler) Preempt(pod *v1.Pod, nodeLister algorithm.NodeLister, err error) (selectedNode *v1.Node, preemptedPods []*v1.Pod, cleanupNominatedPods []*v1.Pod, err error)
