package telemetry

import (
	"sync"
	"time"
)

// FailedAuth counts failed authentications (a missing or unknown token) per
// minute. It keeps only counts: never the token, a digest of it, or anything
// else that could identify who sent it. Only refused requests touch it, so it
// adds no work to successful ones. Counts older than seven days are dropped.
type FailedAuth struct {
	mu      sync.Mutex
	minutes map[int64]int // minute start (Unix ms) → failures in that minute
}

const failedAuthKeep = 7 * 24 * time.Hour

// NewFailedAuth returns an empty counter.
func NewFailedAuth() *FailedAuth { return &FailedAuth{minutes: map[int64]int{}} }

// Add counts one failed authentication at time at.
func (f *FailedAuth) Add(at time.Time) {
	minute := at.Truncate(time.Minute).UnixMilli()
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, seen := f.minutes[minute]; !seen {
		old := at.Add(-failedAuthKeep).UnixMilli()
		for m := range f.minutes {
			if m < old {
				delete(f.minutes, m)
			}
		}
	}
	f.minutes[minute]++
}

// Count returns how many failed authentications happened in minutes starting
// in [from, to], and the start of the first and last such minute (Unix ms).
func (f *FailedAuth) Count(from, to time.Time) (n int, first, last int64) {
	lo, hi := from.Truncate(time.Minute).UnixMilli(), to.UnixMilli()
	f.mu.Lock()
	defer f.mu.Unlock()
	for m, c := range f.minutes {
		if m < lo || m > hi || c == 0 {
			continue
		}
		if n == 0 || m < first {
			first = m
		}
		if m > last {
			last = m
		}
		n += c
	}
	return n, first, last
}
