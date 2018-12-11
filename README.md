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
	//
	// Scheduler stops running the remaining filter functions for a node once one of these filters
	// fails for the node.
	Filter(pod *v1.Pod, nodes [](*v1.Node)) (filteredNodes [](*v1.Node), err error)
}

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
	Score(pod *v1.Pod, nodes [](*v1.Node)) (hostPriorities *schedulerapi.HostPriorityList, weight int, err error)
}
```

.. and others not defined yet.

## Usage

```go
// Define your scheduler plugins
type MyFilter struct {
    // ..
}

func (f *MyFilter) Filter(pod *v1.Pod, nodes [](*v1.Node)) (filteredNodes [](*v1.Node), err error) {
    // ..
}

type MyScorer struct {
    // ..
}

func (s *MyScorer) Score(pod *v1.Pod, nodes [](*v1.Node)) (hostPriorities *schedulerapi.HostPriorityList, weight int, err error) {
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