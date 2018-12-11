package pod

import (
	"sync"

	"k8s.io/api/core/v1"
)

// Map stores a map associating "key" with *v1.Pod.
// It wraps sync.Map for type-safetiness.
type Map struct {
	inner sync.Map
}

// Load returns the pod associated with the key.
// If the pod does not exist, the second return value will be false.
func (m *Map) Load(key string) (*Pod, bool) {
	p, ok := m.inner.Load(key)
	if !ok {
		return nil, false
	}
	pod := p.(Pod)
	return &pod, true
}

// Store stores a new pair of key and pod.
func (m *Map) Store(key string, pod Pod) {
	m.inner.Store(key, pod)
}

// Delete deletes a pod associated with the key.
func (m *Map) Delete(key string) {
	m.inner.Delete(key)
}

// ListPods returns an array of pods.
func (m *Map) ListPods() []*v1.Pod {
	pods := []*v1.Pod{}
	m.Range(func(_ string, pod Pod) bool {
		pods = append(pods, pod.ToV1())
		return true
	})
	return pods
}

// Range applies a function to each pair of key and pod.
func (m *Map) Range(f func(string, Pod) bool) {
	g := func(key, pod interface{}) bool {
		return f(key.(string), pod.(Pod))
	}
	m.inner.Range(g)
}
