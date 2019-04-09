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

package submittermap_test

import (
	"testing"

	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	. "github.com/pfnet-research/k8s-cluster-simulator/pkg/kubesim/submitter_map"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
	"github.com/stretchr/testify/assert"
)

type mySubmitter struct {
	name string
}

func (s *mySubmitter) Submit(
	_ clock.Clock,
	_ algorithm.NodeLister,
	_ metrics.Metrics) ([]submitter.Event, error) {
	return []submitter.Event{}, nil
}

func TestSubmitterMapStore(t *testing.T) {
	submMap := New()

	submMap.Store("foo", &mySubmitter{"foo"})
	actual := submMap.Len()
	expected := 1
	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}

	submMap.Store("bar", &mySubmitter{"bar"})
	actual = submMap.Len()
	expected = 2
	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestSubmitterMapLoad(t *testing.T) {
	submMap := New()

	_, ok := submMap.Load("foo")
	if ok {
		t.Error("got: true\nwant: false")
	}

	subm := mySubmitter{"foo"}
	submMap.Store("foo", &subm)
	actual, ok := submMap.Load("foo")
	if !ok {
		t.Error("got: false\nwant: true")
	}
	if actual.(*mySubmitter).name != subm.name {
		t.Errorf("got: %+v\nwant: %+v", actual, subm)
	}
	if submMap.Len() != 1 {
		t.Errorf("got %+v\nwant: 1", submMap.Len())
	}
}

func TestSubmitterMapDelete(t *testing.T) {
	submMap := New()

	_, ok := submMap.Delete("foo")
	if ok {
		t.Error("got: true\nwant: false")
	}

	subm := mySubmitter{"foo"}

	submMap.Store("foo", &subm)
	if submMap.Len() != 1 {
		t.Errorf("got %+v\nwant: 1", submMap.Len())
	}

	actual, ok := submMap.Delete("foo")
	if !ok {
		t.Error("got: false\nwant: true")
	}
	if actual.(*mySubmitter).name != subm.name {
		t.Errorf("got: %+v\nwant: %+v", actual, subm)
	}
	if submMap.Len() != 0 {
		t.Errorf("got %+v\nwant: 0", submMap.Len())
	}

	_, ok = submMap.Delete("foo")
	if ok {
		t.Error("got: true\nwant: false")
	}
	if submMap.Len() != 0 {
		t.Errorf("got %+v\nwant: 0", submMap.Len())
	}
}

func TestSubmitterMapRange(t *testing.T) {
	submMap := New()

	submMap.Store("foo", &mySubmitter{"foo"})
	submMap.Store("bar", &mySubmitter{"bar"})

	names := make([]string, 0)
	submMap.Range(func(key string, _ submitter.Submitter) bool {
		names = append(names, key)
		return true
	})
	assert.ElementsMatch(t, names, []string{"foo", "bar"})

	names = make([]string, 0)
	submMap.Range(func(key string, _ submitter.Submitter) bool {
		names = append(names, key)
		return false
	})

	assert.Len(t, names, 1)
	if names[0] != "foo" && names[0] != "bar" {
		t.Errorf("got: %+v\nwant: \"foo\" or \"bar\"", names[0])
	}
}
