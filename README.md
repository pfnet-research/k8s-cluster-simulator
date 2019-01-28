# Kubernetes cluster simulator for schedulers

## Usage

See [examples/main.go](examples/main.go).

## Pod submitter and scheduler interface

See [api/submitter.go](api/submitter.go) and [api/scheduler.go](api/scheduler.go).
For the scheduler interface, currently only a subset of the interface is defined.

Note that these interfaces are drafts, subject to change.

## How to specify the resource usage of each pod

Embed a YAML in the annotation field of the pod. e.g.

```yaml
metadata:
  name: nginx-sim
  annotations:
    simSpec: |
- seconds: 5        # a execution phase of this pod
  resourceUsage:    # resource usage (not request nor limit)
    cpu: 1
    memory: 2Gi
    nvidia.com/gpu: 0
- seconds: 10       # another phase that follows the previous one
  resourceUsage:    
    cpu: 2
    memory: 4Gi
    nvidia.com/gpu: 1
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
        Allocatable: // same as Capacity
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
