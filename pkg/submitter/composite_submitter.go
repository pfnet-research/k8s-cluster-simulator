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

package submitter

import (
	"github.com/containerd/containerd/log"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
)

// CompositeSubmitter gathers multiple submitters into a single composite submitter.
type CompositeSubmitter struct {
	submitters map[string]Submitter
}

// NewCompositeSubmitter creates a new CompositeSubmitter with the given submitters.
// Submitters are identified by their name of string.
func NewCompositeSubmitter(submitters map[string]Submitter) *CompositeSubmitter {
	return &CompositeSubmitter{
		submitters,
	}
}

// Submit implements Submitter interface.
// This methods gathers submitter events returned by each submitter.
// Returns an error if a submitter raises an error.
func (c *CompositeSubmitter) Submit(
	clock clock.Clock,
	nodeLister algorithm.NodeLister,
	metrics metrics.Metrics) ([]Event, error) {

	events := []Event{}
	for name, subm := range c.submitters {
		log.L.Debugf("Submitter %s", name)

		ev, err := subm.Submit(clock, nodeLister, metrics)
		if err != nil {
			return []Event{}, err
		}

		for _, e := range ev {
			if _, ok := e.(*TerminateSubmitterEvent); ok {
				delete(c.submitters, name)
				if len(c.submitters) == 0 {
					events = append(events, &TerminateSubmitterEvent{})
				}
			} else {
				events = append(events, e)
			}
		}
	}

	return events, nil
}
