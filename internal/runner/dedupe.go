package runner

import (
	"sync"
)

// DedupeKey returns a string key identifying a call for deduplication.
type DedupeKey func(call Call) string

// DefaultDedupeKey uses service + method + address as the key.
func DefaultDedupeKey(call Call) string {
	return call.Service + "|" + call.Method + "|" + call.Address
}

// DedupeFilter tracks seen calls and skips duplicates within a batch.
type DedupeFilter struct {
	mu   sync.Mutex
	seen map[string]struct{}
	keyFn DedupeKey
}

// NewDedupeFilter creates a new DedupeFilter with the given key function.
// If keyFn is nil, DefaultDedupeKey is used.
func NewDedupeFilter(keyFn DedupeKey) *DedupeFilter {
	if keyFn == nil {
		keyFn = DefaultDedupeKey
	}
	return &DedupeFilter{
		seen:  make(map[string]struct{}),
		keyFn: keyFn,
	}
}

// IsDuplicate returns true if the call has been seen before, and records it if not.
func (d *DedupeFilter) IsDuplicate(call Call) bool {
	key := d.keyFn(call)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.seen[key]; exists {
		return true
	}
	d.seen[key] = struct{}{}
	return false
}

// Reset clears all seen keys.
func (d *DedupeFilter) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]struct{})
}

// SeenCount returns the number of unique calls seen.
func (d *DedupeFilter) SeenCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}
