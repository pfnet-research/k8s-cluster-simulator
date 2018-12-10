package clock

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Clock wraps a time.Time to represent a simulated time.
type Clock struct {
	inner time.Time
}

func NewClock(t time.Time) Clock {
	return Clock{inner: t}
}

func (c Clock) ToMetaV1() metav1.Time {
	return metav1.NewTime(c.inner)
}

func (c Clock) Sub(rhs Clock) time.Duration {
	return c.inner.Sub(rhs.inner)
}

func (c Clock) Add(dur time.Duration) Clock {
	return Clock{inner: c.inner.Add(dur)}
}

func (c Clock) String() string {
	return c.inner.String()
}
