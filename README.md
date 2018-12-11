# Kubernetes cluster simulator for schedulers

## Scheduler interface

NOTE: this interface is a draft, subject to change.

```go
type Filter interface {
	// Filter filters out nodes that cannot run the pod.
	//
	// Scheduler runs filter plugins per node in the same order that they are registered,
	// but scheduler may run these filter function for multiple nodes in parallel.
	// So these plugins must use synchronization when they modify state.
	Filter(pod *v1.Pod, nodes []*v1.Node) (filteredNodes []*v1.Node, err error)
}

// NodeScore represents the score of scheduling to a particular node.
// Higher score means higher priority.
type NodeScore struct {
	// Name of the nodnode.
	Node string
	// Score associated with the node.
	Score int
}

// NodeScoreList declares a []NodeScore type.
type NodeScoreList []NodeScore

type Scorer interface {
	// Score ranks nodes that have passed the filtering stage.
	//
	// Similar to Filter plugins, these are called per node serially in the same order registered,
	// but scheduler may run them for multiple nodes in parallel.
	//
	// Each one of these functions return a score for the given node.
	// The score is multiplied by the weight of the function and aggregated with the result of
	// other scoring functions to yield a total score for the node.
	//
	// These functions can never block scheduling.
	// In case of an error they should return zero for the Node being ranked.
	Score(pod *v1.Pod, nodes []*v1.Node) (scores *NodeScoreList, weight int, err error)
}
```

.. and others not defined yet.

## How to specify the resource usage of each pod

Embed a yaml in the annotation field of the pod. 

e.g.

```yaml
metadata:
  name: nginx-sim
  annotations:
    simSpec: |
- seconds: 5
  resourceUsage:
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10
  resourceUsage:
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1`
```

## Usage

```go
// Define your scheduler plugins
type MyFilter struct {
    // ..
}

func (f *MyFilter) Filter(pod *v1.Pod, nodes []*v1.Node) (filteredNodes []*v1.Node, err error) {
    // ..
}

type MyScorer struct {
    // ..
}

func (s *MyScorer) Score(pod *v1.Pod, nodes []*v1.Node) (scores *NodeScoreList, weight int, err error) {
    // ..
}

func main() {
    configPath := // .. 

    // Create a new kubernetes cluster simulator with the config file path
    kubesim, err := kubesim.NewKubeSim(configPath)
    if err != nil {
        //  ..
    }

    // Register your scheduler plugins
    kubesim.RegisterFilter(MyFilter{})
    kubesim.RegisterScorer(MyScorer{})

    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sig
        cancel()
    }()

    go func() {
        // Continuously submit new pods to the cluster
        for {
            pod := // ..
            kubesim.SubmitPod(pod)
        }
    }()

    // Run the main loop, which invokes scheduler plugins and schedules submitted pods to a selected node
    if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
        // ..
    }
}
```

## Supported `v1.Node` fields

```go
v1.Node{
    TypeMeta: metav1.TypeMeta{
        Kind:       "Node",
        APIVersion: "v1",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:        // all determined by config
        Namespace:   // 
        Labels:      // 
        Annotations: // 
    },
    Spec: v1.NodeSpec{
        Unschedulable: false,
        Taints:        // determined by config
    },
    Status: v1.NodeStatus{
        Capacity:    // determined by config
        Allocatable: // simulated
        Conditions:  []v1.NodeCondition{
            {
                Type:               v1.NodeReady,
                Status:             v1.ConditionTrue,
                LastHeartbeatTime:  // clock
                LastTransitionTime: // clock
                Reason:             "KubeletReady",
                Message:            "kubelet is ready.",
            },
            {
                Type:               "OutOfDisk",
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasSufficientDisk",
                Message:            "kubelet has sufficient disk space available",
            },
            {
                Type:               "MemoryPressure",
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasSufficientMemory",
                Message:            "kubelet has sufficient memory available",
            },
            {
                Type:               "DiskPressure",
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "KubeletHasNoDiskPressure",
                Message:            "kubelet has no disk pressure",
            },
            {
                Type:               "NetworkUnavailable",
                Status:             v1.ConditionFalse,
                LastHeartbeatTime:  // clock,
                LastTransitionTime: // clock,
                Reason:             "RouteCreated",
                Message:            "RouteController created a route",
            },
        },
    },
}
```