package runner

import (
	"context"
	"testing"
	"time"
)

func TestDeadlinePolicy_ZeroDeadline_NoEnforcement(t *testing.T) {
	p := NewDeadlinePolicy(time.Time{})
	ctx, cancel := p.Wrap(context.Background())
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Error("expected no deadline on context")
	}
}

func TestDeadlinePolicy_FutureDeadline_SetsContextDeadline(t *testing.T) {
	deadline := time.Now().Add(5 * time.Second)
	p := NewDeadlinePolicy(deadline)
	ctx, cancel := p.Wrap(context.Background())
	defer cancel()
	d, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	if !d.Equal(deadline) {
		t.Errorf("expected deadline %v, got %v", deadline, d)
	}
}

func TestDeadlinePolicy_Exceeded_PastDeadline(t *testing.T) {
	p := NewDeadlinePolicy(time.Now().Add(-1 * time.Second))
	if !p.Exceeded() {
		t.Error("expected deadline to be exceeded")
	}
}

func TestDeadlinePolicy_NotExceeded_FutureDeadline(t *testing.T) {
	p := NewDeadlinePolicy(time.Now().Add(10 * time.Second))
	if p.Exceeded() {
		t.Error("expected deadline not to be exceeded")
	}
}

func TestDeadlinePolicy_ZeroDeadline_NotExceeded(t *testing.T) {
	p := NewDeadlinePolicy(time.Time{})
	if p.Exceeded() {
		t.Error("zero deadline should never be exceeded")
	}
}

func TestDeadlineFromDuration_Positive(t *testing.T) {
	p := NewDeadlineFromDuration(2 * time.Second)
	if p.deadline.IsZero() {
		t.Error("expected non-zero deadline")
	}
	if p.Remaining() <= 0 {
		t.Error("expected positive remaining time")
	}
}

func TestDeadlineFromDuration_Zero_NoDeadline(t *testing.T) {
	p := NewDeadlineFromDuration(0)
	if !p.deadline.IsZero() {
		t.Error("expected zero deadline for zero duration")
	}
}

func TestDeadlinePolicy_String_NoDeadline(t *testing.T) {
	p := NewDeadlinePolicy(time.Time{})
	if p.String() != "no deadline" {
		t.Errorf("unexpected string: %s", p.String())
	}
}

func TestDeadlinePolicy_String_WithDeadline(t *testing.T) {
	p := NewDeadlinePolicy(time.Now().Add(5 * time.Second))
	s := p.String()
	if s == "no deadline" || s == "" {
		t.Errorf("unexpected string: %s", s)
	}
}
