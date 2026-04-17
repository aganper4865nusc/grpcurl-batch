package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrBulkheadFull is returned when the bulkhead capacity is exceeded.
var ErrBulkheadFull = errors.New("bulkhead: max concurrent calls reached")

// BulkheadPolicy limits concurrent executions and rejects excess calls immediately.
type BulkheadPolicy struct {
	max     int64
	active  atomic.Int64
}

// NewBulkheadPolicy creates a BulkheadPolicy with the given max concurrency.
// If max <= 0, it defaults to 1.
func NewBulkheadPolicy(max int) *BulkheadPolicy {
	if max <= 0 {
		max = 1
	}
	return &BulkheadPolicy{max: int64(max)}
}

// Do executes fn if capacity allows, otherwise returns ErrBulkheadFull immediately.
func (b *BulkheadPolicy) Do(ctx context.Context, fn func(context.Context) error) error {
	current := b.active.Add(1)
	if current > b.max {
		b.active.Add(-1)
		return ErrBulkheadFull
	}
	defer b.active.Add(-1)
	return fn(ctx)
}

// Active returns the number of currently running executions.
func (b *BulkheadPolicy) Active() int64 {
	return b.active.Load()
}

// Max returns the configured maximum concurrency.
func (b *BulkheadPolicy) Max() int64 {
	return b.max
}
