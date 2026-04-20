package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrBudgetExhausted is returned when the error budget is fully consumed.
var ErrBudgetExhausted = errors.New("error budget exhausted")

// BudgetPolicy tracks a rolling error budget and rejects calls once exhausted.
type BudgetPolicy struct {
	max      int64
	used     atomic.Int64
	window   time.Duration
	resetAt  atomic.Int64 // unix nano
}

// NewBudgetPolicy creates a policy allowing at most maxErrors failures per window.
func NewBudgetPolicy(maxErrors int, window time.Duration) *BudgetPolicy {
	if maxErrors <= 0 {
		maxErrors = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	p := &BudgetPolicy{max: int64(maxErrors), window: window}
	p.resetAt.Store(time.Now().Add(window).UnixNano())
	return p
}

func (p *BudgetPolicy) maybeReset() {
	now := time.Now().UnixNano()
	if now >= p.resetAt.Load() {
		p.used.Store(0)
		p.resetAt.Store(time.Now().Add(p.window).UnixNano())
	}
}

// Remaining returns how many errors remain in the current window.
func (p *BudgetPolicy) Remaining() int64 {
	p.maybeReset()
	r := p.max - p.used.Load()
	if r < 0 {
		return 0
	}
	return r
}

// Record consumes one unit of the error budget.
func (p *BudgetPolicy) Record() {
	p.maybeReset()
	p.used.Add(1)
}

// Wrap executes fn, recording a budget unit on failure and blocking if exhausted.
func (p *BudgetPolicy) Wrap(ctx context.Context, fn func(context.Context) error) error {
	p.maybeReset()
	if p.used.Load() >= p.max {
		return ErrBudgetExhausted
	}
	err := fn(ctx)
	if err != nil {
		p.Record()
	}
	return err
}
