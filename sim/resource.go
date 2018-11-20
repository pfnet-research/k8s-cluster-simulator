package sim

import (
	"k8s.io/api/core/v1"
)

func sumResourceList(r1, r2 v1.ResourceList) v1.ResourceList {
	sum := r1
	for r2Key := range r2 {
		if r2Val, ok := sum[r2Key]; ok {
			k := sum[r2Key]
			k.Add(r2Val)
			sum[r2Key] = k
		} else {
			sum[r2Key] = r2Val
		}
	}
	return sum
}

// getResourceRequest extracts total requested resource of the pod.
func getResourceRequest(pod *v1.Pod) v1.ResourceList {
	result := v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		result = sumResourceList(result, container.Resources.Requests)
	}
	return result
}

// greaterEqual returns true when r1 >= r2, false otherwise.
func greaterEqual(r1, r2 v1.ResourceList) bool {
	for r2Key, r2Val := range r2 {
		if r1Val, ok := r1[r2Key]; !ok {
			return false
		} else if r1Val.Cmp(r2Val) <= 0 {
			return false
		}
	}
	return true
}
