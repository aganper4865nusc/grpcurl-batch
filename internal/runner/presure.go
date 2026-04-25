package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrBackpressure is returned when the system is under too much pressure.
var ErrBackpressure = errors.New("backpressure: system overloaded, call rejected")

// BackpressurePolicy rejects new calls when inflight count exceeds a
// high-water mark, and allows them again once it drops below a low-water mark.
type BackpressurePolicy struct {
	highWater int64
	lowWater  int64
	inflight  atomic.Int64
	open      atomic.Bool // true means shedding (backpressure active)
}

// NewBackpressurePolicy creates a policy that activates backpressure when
// inflight calls reach highWater and deactivates when they fall to lowWater.
// If highWater <= 0 it defaults to 50. lowWater defaults to highWater/2.
func NewBackpressurePolicy(highWater, lowWater int) *BackpressurePolicy {
	if highWater <= 0 {
		highWater = 50
	}
	if lowWater <= 0 || lowWater >= highWater {
		lowWater = highWater / 2
	}
	return &BackpressurePolicy{
		highWater: int64(highWater),
		lowWater:  int64(lowWater),
	}
}

// Wrap executes fn under backpressure control.
func (p *BackpressurePolicy) Wrap(ctx context.Context, call Call, fn CallFunc) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	current := p.inflight.Add(1)
	defer func() {
		remaining := p.inflight.Add(-1)
		// Deactivate backpressure once we drain below the low-water mark.
		if p.open.Load() && remaining <= p.lowWater {
			p.open.Store(false)
		}
	}()

	// Activate backpressure when we hit the high-water mark.
	if current >= p.highWater {
		p.open.Store(true)
	}

	if p.open.Load() {
		return Result{}, ErrBackpressure
	}

	return fn(ctx, call)
}

// Inflight returns the current number of in-flight calls.
func (p *BackpressurePolicy) Inflight() int64 {
	return p.inflight.Load()
}

// IsOpen reports whether backpressure is currently active.
func (p *BackpressurePolicy) IsOpen() bool {
	return p.open.Load()
}
