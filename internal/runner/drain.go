package runner

import (
	"context"
	"sync"
	"time"
)

// DrainPolicy waits for in-flight calls to finish before shutdown.
type DrainPolicy struct {
	mu      sync.Mutex
	inflight int
	done    chan struct{}
	timeout time.Duration
}

// NewDrainPolicy creates a DrainPolicy with the given drain timeout.
func NewDrainPolicy(timeout time.Duration) *DrainPolicy {
	return &DrainPolicy{
		done:    make(chan struct{}),
		timeout: timeout,
	}
}

// Acquire marks a call as in-flight.
func (d *DrainPolicy) Acquire() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.inflight++
}

// Release marks a call as finished.
func (d *DrainPolicy) Release() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inflight > 0 {
		d.inflight--
	}
	if d.inflight == 0 {
		select {
		case d.done <- struct{}{}:
		default:
		}
	}
}

// InFlight returns the current number of in-flight calls.
func (d *DrainPolicy) InFlight() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inflight
}

// Wait blocks until all in-flight calls complete or the context is cancelled.
func (d *DrainPolicy) Wait(ctx context.Context) error {
	d.mu.Lock()
	if d.inflight == 0 {
		d.mu.Unlock()
		return nil
	}
	d.mu.Unlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	select {
	case <-d.done:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}
