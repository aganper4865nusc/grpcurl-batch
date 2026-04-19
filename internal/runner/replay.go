package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReplayEntry records a single call execution for later replay.
type ReplayEntry struct {
	Call      Call
	Timestamp time.Time
	Success   bool
}

// ReplayPolicy records failed calls and can replay them.
type ReplayPolicy struct {
	mu      sync.Mutex
	entries []ReplayEntry
	filter  func(ReplayEntry) bool
}

// NewReplayPolicy creates a ReplayPolicy with an optional filter.
// If filter is nil, all failed calls are recorded.
func NewReplayPolicy(filter func(ReplayEntry) bool) *ReplayPolicy {
	if filter == nil {
		filter = func(e ReplayEntry) bool { return !e.Success }
	}
	return &ReplayPolicy{filter: filter}
}

// Record stores a call execution result if it matches the filter.
func (r *ReplayPolicy) Record(call Call, success bool) {
	entry := ReplayEntry{Call: call, Timestamp: time.Now(), Success: success}
	if !r.filter(entry) {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

// Entries returns a snapshot of recorded entries.
func (r *ReplayPolicy) Entries() []ReplayEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ReplayEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// Replay re-executes all recorded calls using the provided executor function.
func (r *ReplayPolicy) Replay(ctx context.Context, exec func(context.Context, Call) error) []error {
	entries := r.Entries()
	var errs []error
	for _, e := range entries {
		if ctx.Err() != nil {
			errs = append(errs, fmt.Errorf("replay cancelled: %w", ctx.Err()))
			break
		}
		if err := exec(ctx, e.Call); err != nil {
			errs = append(errs, fmt.Errorf("replay %s: %w", e.Call.Method, err))
		}
	}
	return errs
}

// Reset clears all recorded entries.
func (r *ReplayPolicy) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = r.entries[:0]
}

// Count returns the number of recorded entries.
func (r *ReplayPolicy) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}
