package clock

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Clock wraps a time.Time to represent a simulated time.
type Clock struct {
	inner time.Time
}

// NewClock creates a new Clock from the time.
func NewClock(t time.Time) Clock {
	return Clock{inner: t}
}

// ToMetaV1 converts this Clock to metav1.Time.
func (c Clock) ToMetaV1() metav1.Time {
	return metav1.NewTime(c.inner)
}

// Add calculates the clock ahead of this Clock by the duration.
func (c Clock) Add(dur time.Duration) Clock {
	return Clock{inner: c.inner.Add(dur)}
}

// Sub calculates the duration from rhs to this Clock.
func (c Clock) Sub(rhs Clock) time.Duration {
	return c.inner.Sub(rhs.inner)
}

// String converts this Clock to a string.
func (c Clock) String() string {
	return c.inner.String()
}
