package runner

import (
	"context"
	"sync"
	"time"
)

// ShadowResult holds the outcome of a shadow call.
type ShadowResult struct {
	Call   Call
	Err    error
	Latency time.Duration
}

// ShadowPolicy duplicates each call to a secondary executor without affecting
// the primary result. Useful for dark-launching new endpoints.
type ShadowPolicy struct {
	mu      sync.Mutex
	shadow  func(ctx context.Context, c Call) error
	results []ShadowResult
	maxResults int
}

// NewShadowPolicy creates a ShadowPolicy that forwards calls to shadowFn
// asynchronously. maxResults caps the in-memory result buffer (0 → 200).
func NewShadowPolicy(shadowFn func(ctx context.Context, c Call) error, maxResults int) *ShadowPolicy {
	if maxResults <= 0 {
		maxResults = 200
	}
	return &ShadowPolicy{
		shadow:     shadowFn,
		maxResults: maxResults,
	}
}

// Wrap executes the primary call normally, then fires the shadow call in a
// detached goroutine so it never blocks or alters the primary outcome.
func (s *ShadowPolicy) Wrap(next func(ctx context.Context, c Call) error) func(ctx context.Context, c Call) error {
	return func(ctx context.Context, c Call) error {
		primaryErr := next(ctx, c)

		// Shadow runs in background; use a detached context so parent
		// cancellation does not abort it.
		go func() {
			start := time.Now()
			sErr := s.shadow(context.Background(), c)
			latency := time.Since(start)

			s.mu.Lock()
			defer s.mu.Unlock()
			if len(s.results) >= s.maxResults {
				// evict oldest
				s.results = s.results[1:]
			}
			s.results = append(s.results, ShadowResult{
				Call:    c,
				Err:     sErr,
				Latency: latency,
			})
		}()

		return primaryErr
	}
}

// Results returns a snapshot of recorded shadow results.
func (s *ShadowPolicy) Results() []ShadowResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ShadowResult, len(s.results))
	copy(out, s.results)
	return out
}

// Reset clears stored shadow results.
func (s *ShadowPolicy) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = s.results[:0]
}
