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

package submittermap

import (
	"sync"

	subm "github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
)

// SubmitterMap wraps sync.Map for type-safetiness.
// The usage of submitters map in KubeSim is read-heavy, so using sync.Map should be more
// performant than using a pair of map and sync.Mutex.
type SubmitterMap struct {
	inner sync.Map
	size  int // track the size to avoid unnecessary computation in Len
}

// New creates a new SubmitterMap.
func New() *SubmitterMap {
	return &SubmitterMap{
		inner: sync.Map{},
		size:  0,
	}
}

// Len returns the number of submitters stored in this SubmitterMap.
func (s *SubmitterMap) Len() int {
	return s.size
}

// Load returns the submitter stored in this SubmitterMap for the key, or nil if no value is
// present.
// The second return value indicates whether the submitter was found.
func (s *SubmitterMap) Load(key string) (subm.Submitter, bool) {
	val, ok := s.inner.Load(key)
	if !ok {
		return nil, ok
	}

	return val.(subm.Submitter), ok
}

// Delete deletes and returns the submitter associated with the key, or returns nil if no value is
// present.
// The second return value indicates whether the submitter was found.
func (s *SubmitterMap) Delete(key string) (subm.Submitter, bool) {
	val, ok := s.Load(key)
	if ok {
		s.inner.Delete(key)
		s.size--
	}

	return val, ok
}

// Store inserts the key and submitter pair to this SubmitterMap.
// If the map had the key, the submitter is updated, and the old submitter is returned.
// The second return value indicates the existence of the old submitter.
func (s *SubmitterMap) Store(key string, submitter subm.Submitter) (subm.Submitter, bool) {
	val, ok := s.Load(key)
	if !ok {
		s.size++
	}
	s.inner.Store(key, submitter)

	return val, ok
}

// Range calls f sequentially for each key and submitter present in this SubmitterMap.
// If f returns false, Range stops the iteration.
func (s *SubmitterMap) Range(f func(key string, submitter subm.Submitter) bool) {
	g := func(key, submitter interface{}) bool {
		return f(key.(string), submitter.(subm.Submitter))
	}

	s.inner.Range(g)
}
