package pod

import (
	"sync"
)

// Map stores a map associating "key"s with Pods.
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

// Store stores a new pair of the key and the pod.
func (m *Map) Store(key string, pod Pod) {
	m.inner.Store(key, pod)
}

// Delete deletes a pod associated with the key.
func (m *Map) Delete(key string) {
	m.inner.Delete(key)
}

// ListPods returns a slice of pods stored in this Map.
func (m *Map) ListPods() []Pod {
	pods := []Pod{}
	m.Range(func(_ string, pod Pod) bool {
		pods = append(pods, pod)
		return true
	})
	return pods
}

// Range applies a function to each pair of a key and a pod.
func (m *Map) Range(f func(string, Pod) bool) {
	g := func(key, pod interface{}) bool {
		return f(key.(string), pod.(Pod))
	}
	m.inner.Range(g)
}
