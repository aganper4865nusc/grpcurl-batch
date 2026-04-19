package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrLoadShed is returned when a call is dropped due to load shedding.
var ErrLoadShed = errors.New("load shed: request dropped")

// ShedderPolicy drops new calls when the number of in-flight requests
// exceeds a configurable threshold.
type ShedderPolicy struct {
	max     int64
	inflight atomic.Int64
}

// NewShedderPolicy creates a ShedderPolicy that sheds calls when
// in-flight count exceeds max. A max <= 0 defaults to 100.
func NewShedderPolicy(max int) *ShedderPolicy {
	if max <= 0 {
		max = 100
	}
	return &ShedderPolicy{max: int64(max)}
}

// Do executes fn if the in-flight count is below the threshold,
// otherwise returns ErrLoadShed immediately.
func (s *ShedderPolicy) Do(ctx context.Context, fn func(context.Context) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	current := s.inflight.Add(1)
	if current > s.max {
		s.inflight.Add(-1)
		return ErrLoadShed
	}
	defer s.inflight.Add(-1)
	return fn(ctx)
}

// Inflight returns the current number of in-flight calls.
func (s *ShedderPolicy) Inflight() int64 {
	return s.inflight.Load()
}
