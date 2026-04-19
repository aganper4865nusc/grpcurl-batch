package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrQuotaExceeded is returned when the call quota has been exhausted.
var ErrQuotaExceeded = errors.New("quota: call quota exceeded")

// QuotaPolicy enforces a maximum number of total calls across the lifetime
// of a batch run.
type QuotaPolicy struct {
	max     int64
	used    atomic.Int64
	disabled bool
}

// NewQuotaPolicy creates a QuotaPolicy with the given maximum. A max <= 0
// disables enforcement (unlimited calls).
func NewQuotaPolicy(max int) *QuotaPolicy {
	q := &QuotaPolicy{max: int64(max)}
	if max <= 0 {
		q.disabled = true
	}
	return q
}

// Allow checks whether another call may proceed. It increments the counter
// atomically and returns ErrQuotaExceeded if the limit is reached.
func (q *QuotaPolicy) Allow() error {
	if q.disabled {
		return nil
	}
	v := q.used.Add(1)
	if v > q.max {
		q.used.Add(-1) // don't count the rejected call
		return ErrQuotaExceeded
	}
	return nil
}

// Wrap executes fn only if the quota allows it.
func (q *QuotaPolicy) Wrap(ctx context.Context, fn func(context.Context) error) error {
	if err := q.Allow(); err != nil {
		return err
	}
	return fn(ctx)
}

// Used returns the number of calls that have consumed quota.
func (q *QuotaPolicy) Used() int64 { return q.used.Load() }

// Remaining returns how many calls are left. Returns -1 when unlimited.
func (q *QuotaPolicy) Remaining() int64 {
	if q.disabled {
		return -1
	}
	r := q.max - q.used.Load()
	if r < 0 {
		return 0
	}
	return r
}

// Reset clears the used counter (useful between test runs).
func (q *QuotaPolicy) Reset() { q.used.Store(0) }
