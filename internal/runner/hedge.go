package runner

import (
	"context"
	"sync"
	"time"
)

// HedgePolicy issues a duplicate call after a delay if the first hasn't completed,
// returning whichever result arrives first.
type HedgePolicy struct {
	delay   time.Duration
	maxHedge int
}

// NewHedgePolicy creates a HedgePolicy. delay is how long to wait before issuing
// a hedge; maxHedge is the maximum number of extra calls (0 disables hedging).
func NewHedgePolicy(delay time.Duration, maxHedge int) *HedgePolicy {
	if maxHedge < 0 {
		maxHedge = 0
	}
	return &HedgePolicy{delay: delay, maxHedge: maxHedge}
}

type hedgeResult struct {
	resp string
	err  error
}

// Execute runs fn and issues up to maxHedge duplicate calls after each delay,
// returning the first successful response or the last error.
func (h *HedgePolicy) Execute(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error) {
	if h.maxHedge == 0 {
		return fn(ctx)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan hedgeResult, h.maxHedge+1)
	var wg sync.WaitGroup

	launch := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := fn(ctx)
			select {
			case ch <- hedgeResult{resp, err}:
			default:
			}
		}()
	}

	launch()

	for i := 0; i < h.maxHedge; i++ {
		select {
		case r := <-ch:
			if r.err == nil {
				return r.resp, nil
			}
		case <-time.After(h.delay):
			launch()
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Wait for first result from remaining goroutines.
	select {
	case r := <-ch:
		return r.resp, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
