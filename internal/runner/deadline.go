package runner

import (
	"context"
	"fmt"
	"time"
)

// DeadlinePolicy enforces an absolute wall-clock deadline across an entire batch run.
type DeadlinePolicy struct {
	deadline time.Time
}

// NewDeadlinePolicy creates a DeadlinePolicy that expires at the given time.
// A zero deadline means no enforcement.
func NewDeadlinePolicy(deadline time.Time) *DeadlinePolicy {
	return &DeadlinePolicy{deadline: deadline}
}

// NewDeadlineFromDuration creates a DeadlinePolicy relative to now.
func NewDeadlineFromDuration(d time.Duration) *DeadlinePolicy {
	if d <= 0 {
		return &DeadlinePolicy{}
	}
	return &DeadlinePolicy{deadline: time.Now().Add(d)}
}

// Wrap returns a derived context that expires at the deadline.
// If the deadline is zero, the parent context is returned unchanged.
func (p *DeadlinePolicy) Wrap(ctx context.Context) (context.Context, context.CancelFunc) {
	if p.deadline.IsZero() {
		return ctx, func() {}
	}
	return context.WithDeadline(ctx, p.deadline)
}

// Remaining returns how much time is left before the deadline.
// Returns 0 if no deadline is set. Returns a negative duration if already exceeded.
func (p *DeadlinePolicy) Remaining() time.Duration {
	if p.deadline.IsZero() {
		return 0
	}
	return time.Until(p.deadline)
}

// Exceeded reports whether the deadline has already passed.
func (p *DeadlinePolicy) Exceeded() bool {
	return !p.deadline.IsZero() && time.Now().After(p.deadline)
}

// IsSet reports whether a deadline has been configured.
func (p *DeadlinePolicy) IsSet() bool {
	return !p.deadline.IsZero()
}

// Deadline returns the underlying deadline time.
// The boolean is false if no deadline is set, matching the context.Context convention.
func (p *DeadlinePolicy) Deadline() (time.Time, bool) {
	return p.deadline, !p.deadline.IsZero()
}

// String returns a human-readable description of the deadline.
func (p *DeadlinePolicy) String() string {
	if p.deadline.IsZero() {
		return "no deadline"
	}
	return fmt.Sprintf("deadline at %s (in %s)", p.deadline.Format(time.RFC3339), p.Remaining().Round(time.Millisecond))
}
