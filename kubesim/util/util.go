package util

import (
	"fmt"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

// TODO: Place these functions at more proper location

// BuildResourceList parses a map from resource names to quantities to v1.ResourceList.
func BuildResourceList(resources map[v1.ResourceName]string) (v1.ResourceList, error) {
	resourceList := v1.ResourceList{}

	for key, value := range resources {
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return nil, strongerrors.InvalidArgument(errors.Errorf("invalid %s value %q", key, value))
		}
		resourceList[key] = quantity
	}

	return resourceList, nil
}

// UpdatePodCondition is copied from "k8s.io/kubernetes/pkg/api/pod".UpdatePodCondition().
// (KubeSim cannot call it because it uses metav1.Now().)
//
// > UpdatePodCondition updates existing pod condition or creates a new one. Sets
// > LastTransitionTime to now if the status has changed. Returns true if pod condition has changed
// > or has been added.
func UpdatePodCondition(clock clock.Clock, status *v1.PodStatus, condition *v1.PodCondition) bool {
	condition.LastTransitionTime = clock.ToMetaV1()
	conditionIndex, oldCondition := podutil.GetPodCondition(status, condition.Type)

	if oldCondition == nil {
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	return !isEqual
}

// PodTotalResourceRequests extracts the total amount of resource requested by this pod.
func PodTotalResourceRequests(pod *v1.Pod) v1.ResourceList {
	result := v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		result = ResourceListSum(result, container.Resources.Requests)
	}
	return result
}

// ResourceListSum returns the sum of two resource lists.
func ResourceListSum(r1, r2 v1.ResourceList) v1.ResourceList {
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

// ErrResourceListDiffNotGE is returned from diffResourceList.
var ErrResourceListDiffNotGE = errors.New("ResourceList is not greater equal")

// ResourceListGE returns true when r1 >= r2, false otherwise.
func ResourceListGE(r1, r2 v1.ResourceList) bool {
	for r2Key, r2Val := range r2 {
		if r1Val, ok := r1[r2Key]; !ok {
			return false
		} else if r1Val.Cmp(r2Val) < 0 {
			return false
		}
	}
	return true
}

// PodKey builds a key for the given pod.
// Returns error if the pod does not have valid (= non-empty) namespace and name.
func PodKey(pod *v1.Pod) (string, error) {
	if pod.ObjectMeta.Namespace == "" {
		return "", strongerrors.InvalidArgument(errors.New("Empty pod namespace"))
	}

	if pod.ObjectMeta.Name == "" {
		return "", strongerrors.InvalidArgument(errors.New("Empty pod name"))
	}

	return PodKeyFromNames(pod.ObjectMeta.Namespace, pod.ObjectMeta.Name), nil
}

// PodKeyFromNames builds a key from the namespace and pod name.
func PodKeyFromNames(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
