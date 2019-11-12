// Copyright 2019 Preferred Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"

	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// BuildResourceList parses a map from resource names to quantities (in strings) to a
// v1.ResourceList.
// Returns error if failed to parse.
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

// PodTotalResourceRequests extracts the total amount of resource requested by the given pod.
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

// ResourceListSub returns the substraction r1-r2
func ResourceListSub(r1, r2 v1.ResourceList) v1.ResourceList {
	sum := r1.DeepCopy()
	for r2Key, r2Val := range r2 {
		if r1Val, ok := sum[r2Key]; ok {
			r1Val.Sub(r2Val)
			sum[r2Key] = r1Val
		} else {
			sum[r2Key] = r2Val
		}
	}
	return sum
}

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

func ResourceListLEWithFactor(r1, r2 v1.ResourceList, factor float64) bool {
	for r2Key, r2Val := range r2 {
		r1Val := r1[r2Key]
		v1 := r1Val.Get()
		v2 := r2Val.Get() * factor
		if v1 > v2 {
			return false
		}
	}
	return true
}

func ResourceListGEWithFactor(r1, r2 v1.ResourceList, factor float64) bool {
	for r2Key, r2Val := range r2 {
		r1Val := r1[r2Key]
		v1 := r1Val.Get()
		v2 := r2Val.Get() * factor
		if v1 < v2 {
			return false
		}
	}
	return true
}

// PodPriority returns the priority of the given pod.
func PodPriority(pod *v1.Pod) int32 {
	prio := int32(scheduling.DefaultPriorityWhenNoDefaultClassExists)
	if pod.Spec.Priority != nil {
		prio = *pod.Spec.Priority
	}
	return prio
}

// PodKey builds a key for the given pod.
// Returns error if the pod doesn't have valid (i.e., non-empty) namespace and name.
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
