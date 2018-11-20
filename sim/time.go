package sim

import (
	"time"
)

// Time wraps time.Time to represent simulated time.
type Time struct {
	time.Time
}

func (t Time) sub(rhs Time) time.Duration {
	return t.Time.Sub(rhs.Time)
}

func (t Time) add(dur time.Duration) Time {
	return Time{t.Time.Add(dur)}
}
