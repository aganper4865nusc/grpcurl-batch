package runner

import (
	"context"
	"sync"
	"time"
)

// ThrottlePolicy limits the rate of calls over a sliding window.
type ThrottlePolicy struct {
	mu       sync.Mutex
	window   time.Duration
	maxCalls int
	timestamps []time.Time
}

// NewThrottlePolicy creates a ThrottlePolicy that allows at most maxCalls
// within the given window duration. Zero or negative maxCalls disables throttling.
func NewThrottlePolicy(maxCalls int, window time.Duration) *ThrottlePolicy {
	return &ThrottlePolicy{
		window:   window,
		maxCalls: maxCalls,
	}
}

// Wait blocks until the call is permitted or ctx is cancelled.
func (t *ThrottlePolicy) Wait(ctx context.Context) error {
	if t.maxCalls <= 0 {
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		t.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-t.window)
		// evict old timestamps
		valid := t.timestamps[:0]
		for _, ts := range t.timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		t.timestamps = valid
		if len(t.timestamps) < t.maxCalls {
			t.timestamps = append(t.timestamps, now)
			t.mu.Unlock()
			return nil
		}
		// calculate when the oldest slot frees up
		oldest := t.timestamps[0]
		waitUntil := oldest.Add(t.window)
		t.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(waitUntil)):
		}
	}
}

// Count returns the number of calls recorded in the current window.
func (t *ThrottlePolicy) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	cutoff := time.Now().Add(-t.window)
	count := 0
	for _, ts := range t.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}
