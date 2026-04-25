package runner

import (
	"context"
	"fmt"
	"sync/atomic"
)

// InflightTracker tracks the number of currently in-flight calls and exposes
// a middleware that rejects new calls once a configured maximum is reached.
type InflightTracker struct {
	max     int64
	current atomic.Int64
}

// NewInflightTracker creates a new InflightTracker. If max is <= 0 it defaults
// to 100.
func NewInflightTracker(max int) *InflightTracker {
	if max <= 0 {
		max = 100
	}
	return &InflightTracker{max: int64(max)}
}

// Current returns the number of calls currently in-flight.
func (t *InflightTracker) Current() int64 {
	return t.current.Load()
}

// Acquire increments the in-flight counter if capacity is available.
// It returns an error and a no-op release func when the limit is exceeded.
func (t *InflightTracker) Acquire(ctx context.Context) (release func(), err error) {
	if ctx.Err() != nil {
		return func() {}, ctx.Err()
	}
	for {
		cur := t.current.Load()
		if cur >= t.max {
			return func() {}, fmt.Errorf("inflight limit reached (%d/%d)", cur, t.max)
		}
		if t.current.CompareAndSwap(cur, cur+1) {
			return func() { t.current.Add(-1) }, nil
		}
	}
}

// WithInflightLimit returns a middleware that enforces an in-flight call limit
// using the provided InflightTracker.
func WithInflightLimit(tracker *InflightTracker) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (Result, error) {
			release, err := tracker.Acquire(ctx)
			if err != nil {
				return Result{Call: call, Err: err}, err
			}
			defer release()
			return next(ctx, call)
		}
	}
}

// InflightSnapshot holds a point-in-time view of tracker state.
type InflightSnapshot struct {
	Current int64 `json:"current"`
	Max     int64 `json:"max"`
}

// Snapshot returns the current state of the tracker.
func (t *InflightTracker) Snapshot() InflightSnapshot {
	return InflightSnapshot{
		Current: t.current.Load(),
		Max:     t.max,
	}
}
