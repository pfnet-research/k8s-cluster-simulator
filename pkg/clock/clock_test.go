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

	"github.com/ordovicia/k8s-cluster-simulator/pkg/clock"
)

func TestClockToMetaV1(t *testing.T) {
	now := time.Now()

	actual := metav1.NewTime(now)
	expected := clock.NewClock(now).ToMetaV1()

	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestClockAdd(t *testing.T) {
	now := time.Now()
	clk := clock.NewClock(now)

	dur, _ := time.ParseDuration("12h30m15s")
	actual := clk.Add(dur)
	expected := clock.NewClock(now.Add(dur))

	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestClockSub(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2018-01-01T12:30:15+09:00")
	clk := clock.NewClock(tm)

	tm2, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00+09:00")
	clk2 := clock.NewClock(tm2)

	actual := clk.Sub(clk2)
	expected, _ := time.ParseDuration("12h30m15s")

	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}
