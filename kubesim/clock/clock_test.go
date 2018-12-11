package clock_test

import (
	"testing"
	"time"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
)

func TestClockAdd(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00+09:00")
	clk := clock.NewClock(tm)

	dur, _ := time.ParseDuration("12h30m15s")
	actual := clk.Add(dur)

	tm2, _ := time.Parse(time.RFC3339, "2018-01-01T12:30:15+09:00")
	expected := clock.NewClock(tm2)

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
