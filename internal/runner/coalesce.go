package runner

import (
	"context"
	"sync"
)

// coalesceEntry holds the in-flight result for a deduplicated key.
type coalesceEntry struct {
	wg  sync.WaitGroup
	res *CallResult
	err error
}

// CoalescePolicy merges concurrent identical calls so that only one
// in-flight execution runs; all callers sharing the same key receive
// the same result once it completes.
type CoalescePolicy struct {
	mu      sync.Mutex
	inflight map[string]*coalesceEntry
	keyFn   func(Call) string
}

// NewCoalescePolicy creates a CoalescePolicy. keyFn derives the
// deduplication key from a Call; DefaultDedupeKey is a sensible default.
func NewCoalescePolicy(keyFn func(Call) string) *CoalescePolicy {
	if keyFn == nil {
		keyFn = DefaultDedupeKey
	}
	return &CoalescePolicy{
		inflight: make(map[string]*coalesceEntry),
		keyFn:   keyFn,
	}
}

// Wrap returns a CallFunc that coalesces concurrent identical calls.
func (c *CoalescePolicy) Wrap(next CallFunc) CallFunc {
	return func(ctx context.Context, call Call) (*CallResult, error) {
		key := c.keyFn(call)

		c.mu.Lock()
		if entry, ok := c.inflight[key]; ok {
			c.mu.Unlock()
			// Wait for the in-flight call to finish.
			entry.wg.Wait()
			return entry.res, entry.err
		}

		entry := &coalesceEntry{}
		entry.wg.Add(1)
		c.inflight[key] = entry
		c.mu.Unlock()

		entry.res, entry.err = next(ctx, call)
		entry.wg.Done()

		c.mu.Lock()
		delete(c.inflight, key)
		c.mu.Unlock()

		return entry.res, entry.err
	}
}

// Inflight returns the number of currently in-flight unique keys.
func (c *CoalescePolicy) Inflight() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.inflight)
}
