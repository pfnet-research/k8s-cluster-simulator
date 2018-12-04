package sim

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Time wraps time.Time to represent simulated time.
type Time struct {
	inner time.Time
}

// NewTime resunts a new Time
func NewTime(t time.Time) Time {
	return Time{inner: t}
}

// ToMetaV1 returns a time converted into metav1.Time type
func (t Time) ToMetaV1() metav1.Time {
	return metav1.NewTime(t.inner)
}

func (t Time) Sub(rhs Time) time.Duration {
	return t.inner.Sub(rhs.inner)
}

func (t Time) Add(dur time.Duration) Time {
	return Time{inner: t.inner.Add(dur)}
}

func (t Time) String() string {
	return t.inner.String()
}
