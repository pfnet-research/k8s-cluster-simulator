package node

import (
	"errors"

	"k8s.io/api/core/v1"
)

// sumResourceList returns the sum of two resource lists.
func sumResourceList(r1, r2 v1.ResourceList) v1.ResourceList {
	sum := r1
	for r2Key := range r2 {
		if r2Val, ok := sum[r2Key]; ok {
			r1Val := sum[r2Key]
			r1Val.Add(r2Val)
			sum[r2Key] = r1Val
		} else {
			sum[r2Key] = r2Val
		}
	}
	return sum
}

// errDiffResourceNotGE may be returned from diffResourceList().
var errDiffResourceNotGE = errors.New("resource list is not greater equal")

// diffResourceList returns a difference between two resource lists.
// r1 must be greater or equal than r2, otherwise errDiffResourceNotGe will be returned.
func diffResourceList(r1, r2 v1.ResourceList) (v1.ResourceList, error) {
	if !greaterEqual(r1, r2) {
		return v1.ResourceList{}, errDiffResourceNotGE
	}

	diff := r1
	for r2Key := range r2 {
		r2Val, _ := diff[r2Key]
		r1Val := diff[r2Key]
		r1Val.Sub(r2Val)
		diff[r2Key] = r1Val
	}
	return diff, nil
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
