package node

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

// resourceListSum returns the sum of two resource lists.
func resourceListSum(r1, r2 v1.ResourceList) v1.ResourceList {
	sum := r1.DeepCopy()
	for r2Key, r2Val := range r2 {
		if r1Val, ok := sum[r2Key]; ok {
			r1Val.Add(r2Val)
			sum[r2Key] = r1Val
		} else {
			sum[r2Key] = r2Val
		}
	}
	return sum
}

// errResourceListDiffNotGE is returned from diffResourceList.
var errResourceListDiffNotGE = errors.New("resource list is not greater equal")

// resourceListDiff returns a difference between two resource lists.
// r1 must be greater or equal than r2, otherwise errResourceListDiffNotGE will be returned.
func resourceListDiff(r1, r2 v1.ResourceList) (v1.ResourceList, error) {
	if !resourceListGE(r1, r2) {
		return v1.ResourceList{}, errResourceListDiffNotGE
	}

	diff := r1.DeepCopy()
	for r2Key, r2Val := range r2 {
		r1Val := diff[r2Key]
		r1Val.Sub(r2Val)
		diff[r2Key] = r1Val
	}
	return diff, nil
}

// getResourceReq extracts total requested resource of the pod.
func getResourceReq(pod *v1.Pod) v1.ResourceList {
	result := v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		result = resourceListSum(result, container.Resources.Requests)
	}
	return result
}

// resourceListGE returns true when r1 >= r2, false otherwise.
func resourceListGE(r1, r2 v1.ResourceList) bool {
	for r2Key, r2Val := range r2 {
		if r1Val, ok := r1[r2Key]; !ok {
			return false
		} else if r1Val.Cmp(r2Val) < 0 {
			return false
		}
	}
	return true
}
