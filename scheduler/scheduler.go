package scheduler

import (
	"k8s.io/api/core/v1"
)

type Scheduler interface {
	CreatePod(pod *v1.Pod)
}
