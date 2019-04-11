# Kubernetes cluster simulator

[![GoDoc][godoc-image]][godoc-link]
[![Build Status][build-image]][build-link]
[![Coverage Status][cov-image]][cov-link]

Kubernetes cluster simulator for evaluating schedulers.

## Usage

See [example](example) directory.

```go
// 1. Create a KubeSim with a pod queue and a scheduler.
queue := queue.NewPriorityQueue()
sched := buildScheduler() // see below
kubesim := kubesim.NewKubeSimFromConfigPathOrDie(configPath, queue, sched)

// 2. Register one or more pod submitters to KubeSim.
numOfSubmittingPods := 8
kubesim.AddSubmitter(newMySubmitter(numOfSubmittingPods))

// 3. Run the main loop of KubeSim.
//    In each execution of the loop, KubeSim
//      1) stores pods submitted from the registered submitters to its queue,
//      2) invokes scheduler with pending pods and cluster state,
//      3) emits cluster metrics to designated location(s) if enabled
//      4) progresses the simulated clock
if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
    log.L.Fatal(err)
}

func buildScheduler() scheduler.Scheduler {
    // 1. Create a generic scheduler that mimics a kube-scheduler.
    sched := scheduler.NewGenericScheduler( /* preemption enabled */ true)

    // 2. Register extender(s)
    sched.AddExtender(
        scheduler.Extender{
            Name:             "MyExtender",
            Filter:           filterExtender,
            Prioritize:       prioritizeExtender,
            Weight:           1,
            NodeCacheCapable: true,
        },
    )

    // 2. Register plugin(s)
    // Predicate
    sched.AddPredicate("GeneralPredicates", predicates.GeneralPredicates)
    // Prioritizer
    sched.AddPrioritizer(priorities.PriorityConfig{
        Name:   "BalancedResourceAllocation",
        Map:    priorities.BalancedResourceAllocationMap,
        Reduce: nil,
        Weight: 1,
    })
    sched.AddPrioritizer(priorities.PriorityConfig{
        Name:   "LeastRequested",
        Map:    priorities.LeastRequestedPriorityMap,
        Reduce: nil,
        Weight: 1,
    })

    return &sched
}
```

### Pod submitter interface

See [pkg/submitter/submitter.go](pkg/submitter/submitter.go).

```go
// Submitter defines the submitter interface.
type Submitter interface {
	// Submit submits pods to a simulated cluster.
	// The return value is a list of submitter events.
	// Submitters are called serially in the same order that they are registered to the simulated
	// cluster.
	// This method must never block.
    Submit(clock clock.Clock, nodeLister algorithm.NodeLister, metrics metrics.Metrics) ([]Event, error)
}

// Event defines the interface of a submitter event.
// Submit can returns any type in a list that implements this interface.
type Event interface {
	IsSubmitterEvent() bool
}

// SubmitEvent represents an event of submitting a pod to a cluster.
type SubmitEvent struct {
	Pod *v1.Pod
}

// DeleteEvent represents an event of deleting a pod from a cluster.
type DeleteEvent struct {
	PodName      string
	PodNamespace string
}

// UpdateEvent represents an event of updating the manifest of a pending pod.
type UpdateEvent struct {
	PodName      string
	PodNamespace string
	NewPod       *v1.Pod
}

// TerminateSubmitterEvent represents an event of terminating the submission process.
type TerminateSubmitterEvent struct {
}
```

### `kube-scheduler`-compatible scheduler interface

See [pkg/scheduler/generic_scheduler.go](pkg/scheduler/generic_scheduler.go) and
[pkg/scheduler/extender.go](pkg/scheduler/extender.go).

k8s-cluster-simulator provides `GenericScheduler`, which follows the behavior of kube-scheduler's
`genericScheduler`.
`GenericScheduler` makes scheduling decision for each given pod in the one-by-one manner, with
predicates and prioritizers.

The interfaces of predicates and prioritizers are similar to those of kube-scheduler.

```go
// NewGenericScheduler creates a new GenericScheduler.
func NewGenericScheduler(preeptionEnabled bool) GenericScheduler {
	return GenericScheduler{
		predicates:        map[string]predicates.FitPredicate{},
		preemptionEnabled: preeptionEnabled,
	}
}

// AddExtender adds an extender to this GenericScheduler.
func (sched *GenericScheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// AddPredicate adds a predicate plugin to this GenericScheduler.
func (sched *GenericScheduler) AddPredicate(name string, predicate predicates.FitPredicate) {
	sched.predicates[name] = predicate
}

// AddPrioritizer adds a prioritizer plugin to this GenericScheduler.
func (sched *GenericScheduler) AddPrioritizer(prioritizer priorities.PriorityConfig) {
	sched.prioritizers = append(sched.prioritizers, prioritizer)
}

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
```

### Lowest-level scheduler interface

See [pkg/scheduler/scheduler.go](pkg/scheduler/scheduler.go).

k8s-cluster-simulator also supports the lowest-level scheduler interface, which makes scheduling
decisions for (subset of) pending pods and running pods, given the cluster state at a clock.

```go
// Scheduler defines the lowest-level scheduler interface.
type Scheduler interface {
	// Schedule makes scheduling decisions for (subset of) pending pods and running pods.
	// The return value is a list of scheduling events.
	// This method must never block.
	Schedule(
		clock clock.Clock,
		podQueue queue.PodQueue,
		nodeLister algorithm.NodeLister,
		nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error)
}

// Event defines the interface of a scheduling event.
// Submit can returns any type in a list that implements this interface.
type Event interface {
	IsSchedulerEvent() bool
}

// BindEvent represents an event of deciding the binding of a pod to a node.
type BindEvent struct {
	Pod            *v1.Pod
	ScheduleResult core.ScheduleResult
}

// DeleteEvent represents an event of the deleting a bound pod on a node.
type DeleteEvent struct {
	PodNamespace string
	PodName      string
	NodeName     string
}
```

### How to specify the resource usage of each pod

Embed a YAML in the `annotations` field of the pod manifest. e.g.,

```yaml
metadata:
  name: nginx-sim
  annotations:
    simSpec: |
- seconds: 5        # an execution phase of this pod
  resourceUsage:    # resource usage (not request, nor limit)
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10       # another phase that follows the previous one
  resourceUsage:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
```

## Supported `v1.Pod` fields

These fields are populated or used by the simulator.

```go
v1.Pod{
    ObjectMeta: metav1.ObjectMeta{
        UID,                // populated when this pod is submitted to the simulator
        CreationTimestamp,  // populated when this pod is submitted to the simulator
        DeletionTimestamp,  // populated when a deletion event for this pod has been accepted by the simulator
    },
    Spec: v1.PodSpec {
        NodeName,                       // populated when the cluster binds this pod to a node
        TerminationGracePeriodSeconds,  // read when this pod is deleted
        Priority,                       // read by PriorityQueue to sort pods,
                                        // and read when the scheduler trys to schedule this pod
    },
    Status: v1.PodStatus{
        Phase,              // populated by the simulator. Pending -> Running -> Succeeded xor Failed
        Conditions,         // populated by the simulator
        Reason,             // populated by the simulator
        Message,            // populated by the simulator
        StartTime,          // populated by the simulator when this pod has started its execution
        ContainerStatuses,  // populated by the simulator
    },
}
```

## Supported `v1.Node` fields

These fields are populated and used by the simulator.

```go
v1.Node{
    TypeMeta: metav1.TypeMeta{
        Kind:       "Node",
        APIVersion: "v1",
    },
    ObjectMeta: // determined by the config
    Spec:       // determined by the config
    Status: v1.NodeStatus{
        Capacity:                           // Determined by the config
        Allocatable:                        // Same as Capacity
        Conditions:  []v1.NodeCondition{    // populated by the simulator
            {
                Type:               v1.NodeReady,
                Status:             v1.ConditionTrue,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletReady",
                Message:            "kubelet is posting ready status",
            },
            {
                Type:               v1.NodeOutOfDisk,
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasSufficientDisk",
                Message:            "kubelet has sufficient disk space available",
            },
            {
                Type:               v1.NodeMemoryPressure,
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasSufficientMemory",
                Message:            "kubelet has sufficient memory available",
            },
            {
                Type:               v1.NodeDiskPressure,
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasNoDiskPressure",
                Message:            "kubelet has no disk pressure",
            },
            {
                Type:               v1.NodePIDPressure,
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasSufficientPID",
                Message:            "kubelet has sufficient PID available",
            },
        },
    },
}
```

## Related projects

The design and implementation of this project are inherently inspired by
[kubernetes](https://github.com/kubernetes/kubernetes), which is licensed under Apache-2.0.
Moreover, functions in the following files were obtained from Kubernetes project and modified so
that they would be compatible with k8s-cluster-simulator.
Please see each file for more detail.

* [pkg/scheduler/generic_scheduler_k8s.go](pkg/scheduler/generic_scheduler_k8s.go)
* [pkg/queue/priority_queue_k8s.go](pkg/queue/priority_queue_k8s.go)
* [pkg/util/util_k8s.go](pkg/util/util_k8s.go)


[build-image]: https://travis-ci.com/pfnet-research/k8s-cluster-simulator.svg
[build-link]:  http://travis-ci.com/pfnet-research/k8s-cluster-simulator
[cov-image]:   https://coveralls.io/repos/github/pfnet-research/k8s-cluster-simulator/badge.svg?branch=master
[cov-link]:    https://coveralls.io/github/pfnet-research/k8s-cluster-simulator?branch=master
[godoc-image]: https://godoc.org/github.com/pfnet-research/k8s-cluster-simulator/pkg?status.svg
[godoc-link]:  https://godoc.org/github.com/pfnet-research/k8s-cluster-simulator/pkg
