package runner

import (
	"context"
	"sync"
	"time"
)

// AdaptiveThrottle adjusts its concurrency limit dynamically based on observed
// error rate over a sliding window. When the error rate exceeds HighWatermark
// the limit is halved; when it drops below LowWatermark the limit grows by one
// (up to MaxConcurrency).
type AdaptiveThrottle struct {
	mu             sync.Mutex
	current        int
	max            int
	highWatermark  float64
	lowWatermark   float64
	window         *WindowPolicy
	adjustInterval time.Duration
	stopCh         chan struct{}
}

// NewAdaptiveThrottle creates an AdaptiveThrottle with sensible defaults.
// max is the ceiling concurrency; initial concurrency starts at max.
func NewAdaptiveThrottle(max int, high, low float64, window *WindowPolicy) *AdaptiveThrottle {
	if max < 1 {
		max = 1
	}
	at := &AdaptiveThrottle{
		current:        max,
		max:            max,
		highWatermark:  high,
		lowWatermark:   low,
		window:         window,
		adjustInterval: 5 * time.Second,
		stopCh:         make(chan struct{}),
	}
	go at.adjustLoop()
	return at
}

// Current returns the current concurrency limit.
func (a *AdaptiveThrottle) Current() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.current
}

// Stop halts the background adjustment goroutine.
func (a *AdaptiveThrottle) Stop() {
	close(a.stopCh)
}

func (a *AdaptiveThrottle) adjustLoop() {
	ticker := time.NewTicker(a.adjustInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.adjust()
		case <-a.stopCh:
			return
		}
	}
}

func (a *AdaptiveThrottle) adjust() {
	stats := a.window.Stats()
	if stats.Total == 0 {
		return
	}
	errRate := float64(stats.Failures) / float64(stats.Total)
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case errRate >= a.highWatermark:
		if a.current > 1 {
			a.current /= 2
		}
	case errRate < a.lowWatermark:
		if a.current < a.max {
			a.current++
		}
	}
}

// Wrap executes fn respecting the current adaptive limit via a Semaphore.
func (a *AdaptiveThrottle) Wrap(ctx context.Context, fn func(context.Context) error) error {
	a.mu.Lock()
	limit := a.current
	a.mu.Unlock()
	sem := NewSemaphore(limit)
	if err := sem.Acquire(ctx); err != nil {
		return err
	}
	defer sem.Release()
	return fn(ctx)
}
