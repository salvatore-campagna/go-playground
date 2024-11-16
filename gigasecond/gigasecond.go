// Package gigasecond provides a utility to work with time calculations,
// specifically adding a gigasecond (1 billion seconds) to a given time.
package gigasecond

import "time"

// AddGigasecond takes a time value and returns a new time value that is
// exactly 1 gigasecond (1 billion seconds) later.
func AddGigasecond(t time.Time) time.Time {
	return t.Add(time.Second * 1_000_000_000)
}
