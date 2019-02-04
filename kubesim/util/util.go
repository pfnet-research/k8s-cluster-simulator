package util

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

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
