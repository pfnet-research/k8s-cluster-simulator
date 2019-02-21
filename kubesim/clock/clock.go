package clock

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Clock wraps a time.Time to represent a simulated time.
type Clock struct {
	inner metav1.Time
}

// NewClock creates a new Clock from the time.Time.
func NewClock(t time.Time) Clock {
	return NewClockWithMetaV1(metav1.NewTime(t))
}

// NewClockWithMetaV1 creates a new Clock from the metav1.Time.
func NewClockWithMetaV1(t metav1.Time) Clock {
	return Clock{inner: t}
}

// ToMetaV1 converts this Clock to metav1.Time.
func (c Clock) ToMetaV1() metav1.Time {
	return c.inner
}

// Add calculates the clock ahead of this Clock by the duration.
func (c Clock) Add(dur time.Duration) Clock {
	t := metav1.NewTime(c.inner.Time.Add(dur))
	return NewClockWithMetaV1(t)
}

// Sub calculates the duration from rhs to this Clock.
func (c Clock) Sub(rhs Clock) time.Duration {
	return c.inner.Time.Sub(rhs.inner.Time)
}

// Before returns whether tis Clock is before rhs.
func (c Clock) Before(rhs Clock) bool {
	return c.inner.Before(&rhs.inner)
}

// String converts this Clock to a string.
func (c Clock) String() string {
	return c.inner.String()
}
