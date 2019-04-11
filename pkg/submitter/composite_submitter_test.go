package submitter_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/submitter"
)

type mySubmitter struct {
	fuse  int // returns TerminateSubmitterEvent when cycle reaches fuse
	cycle int
}

func (s *mySubmitter) Submit(
	_ clock.Clock,
	_ algorithm.NodeLister,
	_ metrics.Metrics) ([]submitter.Event, error) {

	if s.cycle == s.fuse {
		return []submitter.Event{&submitter.TerminateSubmitterEvent{}}, nil
	} else {
		s.cycle++
		return []submitter.Event{&submitter.SubmitEvent{}}, nil
	}
}

type myDeleter struct {
	fuse  int // returns error when cycle reaches fuse
	cycle int
}

func (d *myDeleter) Submit(
	_ clock.Clock,
	_ algorithm.NodeLister,
	_ metrics.Metrics) ([]submitter.Event, error) {

	if d.cycle == d.fuse {
		return []submitter.Event{}, errors.New("fused")
	} else {
		d.cycle++
		return []submitter.Event{&submitter.DeleteEvent{}}, nil
	}
}

type dummyNodeLister struct{}

func (d *dummyNodeLister) List() ([]*v1.Node, error) { return []*v1.Node{}, nil }

func TestCompositeSubmitterNoSubmitter(t *testing.T) {
	now := clock.NewClock(time.Now())
	nodeLister := &dummyNodeLister{}
	met := metrics.Metrics{}

	submitters := make(map[string]submitter.Submitter)
	composite := submitter.NewCompositeSubmitter(submitters)

	events, err := composite.Submit(now, nodeLister, met)
	assert.NoError(t, err)
	assert.Empty(t, events)
}

func TestCompositeSubmitterMultiSubmitter(t *testing.T) {
	now := clock.NewClock(time.Now())
	nodeLister := &dummyNodeLister{}
	met := metrics.Metrics{}

	fuse := 4
	submitters := map[string]submitter.Submitter{
		"submitter0": &mySubmitter{fuse: fuse},
		"submitter1": &mySubmitter{fuse: fuse + 1},
	}
	composite := submitter.NewCompositeSubmitter(submitters)

	for i := 0; i < fuse; i++ {
		events, err := composite.Submit(now, nodeLister, met)
		assert.NoError(t, err)
		assert.ElementsMatch(t, events, []submitter.Event{&submitter.SubmitEvent{}, &submitter.SubmitEvent{}})
	}

	// submitter0 terminated
	events, err := composite.Submit(now, nodeLister, met)
	assert.NoError(t, err)
	assert.ElementsMatch(t, events, []submitter.Event{&submitter.SubmitEvent{}})

	// submitter1 terminated
	events, err = composite.Submit(now, nodeLister, met)
	assert.NoError(t, err)
	assert.ElementsMatch(t, events, []submitter.Event{&submitter.TerminateSubmitterEvent{}})
}

func TestCompositeSubmitterError(t *testing.T) {
	now := clock.NewClock(time.Now())
	nodeLister := &dummyNodeLister{}
	met := metrics.Metrics{}

	fuse := 1
	submitters := map[string]submitter.Submitter{
		"submitter0": &myDeleter{fuse: fuse},
	}
	composite := submitter.NewCompositeSubmitter(submitters)

	for i := 0; i < fuse; i++ {
		events, err := composite.Submit(now, nodeLister, met)
		assert.NoError(t, err)
		assert.ElementsMatch(t, events, []submitter.Event{&submitter.DeleteEvent{}})
	}

	_, err := composite.Submit(now, nodeLister, met)
	assert.Error(t, err)
}
