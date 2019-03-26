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

package clock

import (
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Clock wraps a time.Time to represent a simulated time.
type Clock struct {
	inner metav1.Time
}

// NewClock creates a new clock from the time.Time.
func NewClock(t time.Time) Clock {
	return NewClockWithMetaV1(metav1.NewTime(t))
}

// NewClockWithMetaV1 creates a new clock from the metav1.Time.
func NewClockWithMetaV1(t metav1.Time) Clock {
	return Clock{inner: t}
}

// ToMetaV1 converts this clock to metav1.Time.
func (c Clock) ToMetaV1() metav1.Time {
	return c.inner
}

// Add calculates the clock ahead of this clock by the duration.
func (c Clock) Add(dur time.Duration) Clock {
	t := metav1.NewTime(c.inner.Time.Add(dur))
	return NewClockWithMetaV1(t)
}

// Sub calculates the duration from rhs to this clock.
func (c Clock) Sub(rhs Clock) time.Duration {
	return c.inner.Time.Sub(rhs.inner.Time)
}

// Before returns whether this clock is before rhs.
func (c Clock) Before(rhs Clock) bool {
	return c.inner.Before(&rhs.inner)
}

// String converts this clock to a string.
func (c Clock) String() string {
	return c.inner.String()
}

// ToRFC3339 formats this clock to a string in RFC3339 format.
func (c Clock) ToRFC3339() string {
	return c.inner.Format(time.RFC3339)
}

// MarshalJSON implements json.Marshaler.
func (c Clock) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToRFC3339())
}
