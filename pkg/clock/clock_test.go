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

package clock_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
)

func TestClockNewClockAndToMetaV1(t *testing.T) {
	stdTime := time.Now()
	metaV1Time := metav1.NewTime(stdTime)

	simTimeFromStd := clock.NewClock(stdTime)
	simTimeFromMetaV1 := clock.NewClockWithMetaV1(metaV1Time)

	expected := simTimeFromStd.ToMetaV1()
	if metaV1Time != expected {
		t.Errorf("got: %+v\nwant: %+v", metaV1Time, expected)
	}

	expected = simTimeFromMetaV1.ToMetaV1()
	if metaV1Time != expected {
		t.Errorf("got: %+v\nwant: %+v", metaV1Time, expected)
	}
}

func TestClockAdd(t *testing.T) {
	now := time.Now()
	clk := clock.NewClock(now)

	dur, _ := time.ParseDuration("12h30m15s")
	actual := clk.Add(dur)
	expected := clock.NewClock(now.Add(dur))

	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestClockSub(t *testing.T) {
	time0, _ := time.Parse(time.RFC3339, "2018-01-01T12:30:15+09:00")
	clock0 := clock.NewClock(time0)

	time1, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00+09:00")
	clock1 := clock.NewClock(time1)

	actual := clock0.Sub(clock1)
	expected, _ := time.ParseDuration("12h30m15s")

	if actual != expected {
		t.Errorf("got: %+v\nwant: %+v", actual, expected)
	}
}

func TestClockBefore(t *testing.T) {
	time0, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00+09:00")
	clock0 := clock.NewClock(time0)

	time1, _ := time.Parse(time.RFC3339, "2018-01-01T12:30:15+09:00")
	clock1 := clock.NewClock(time1)

	if !clock0.Before(clock1) {
		t.Errorf("got: true\nwant: false")
	}
	if clock1.Before(clock0) {
		t.Errorf("got: false\nwant: true")
	}
}
