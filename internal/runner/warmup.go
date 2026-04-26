package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WarmupPolicy executes a set of probe calls before the main batch begins,
// ensuring downstream services are ready to handle traffic.
type WarmupPolicy struct {
	mu       sync.Mutex
	warmed   bool
	attempts int
	delay    time.Duration
	probe    func(ctx context.Context) error
}

// NewWarmupPolicy creates a WarmupPolicy that will call probe up to attempts
// times with delay between each attempt before declaring the target warm.
func NewWarmupPolicy(probe func(ctx context.Context) error, attempts int, delay time.Duration) *WarmupPolicy {
	if attempts <= 0 {
		attempts = 3
	}
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}
	return &WarmupPolicy{
		attempts: attempts,
		delay:    delay,
		probe:    probe,
	}
}

// Warm runs the probe until it succeeds or attempts are exhausted.
// Subsequent calls to Warm are no-ops once warmed.
func (w *WarmupPolicy) Warm(ctx context.Context) error {
	w.mu.Lock()
	if w.warmed {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	for i := 0; i < w.attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.probe(ctx); err == nil {
			w.mu.Lock()
			w.warmed = true
			w.mu.Unlock()
			return nil
		}
		if i < w.attempts-1 {
			select {
			case <-time.After(w.delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("warmup: target not ready after %d attempts", w.attempts)
}

// IsWarmed reports whether the target has been successfully warmed.
func (w *WarmupPolicy) IsWarmed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.warmed
}

// Reset clears the warmed state so Warm can be called again.
func (w *WarmupPolicy) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.warmed = false
}
