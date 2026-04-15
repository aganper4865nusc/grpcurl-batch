package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errFake = errors.New("fake error")

func TestRetryPolicy_SuccessOnFirstAttempt(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := p.Execute(context.Background(), func(_ context.Context, _ int) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryPolicy_SuccessAfterRetry(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := p.Execute(context.Background(), func(_ context.Context, _ int) error {
		calls++
		if calls < 3 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryPolicy_ExhaustsAttempts(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 2, Delay: 0}
	calls := 0
	err := p.Execute(context.Background(), func(_ context.Context, _ int) error {
		calls++
		return errFake
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestRetryPolicy_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	p := RetryPolicy{MaxAttempts: 3, Delay: time.Second}
	err := p.Execute(ctx, func(_ context.Context, _ int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestRetryPolicy_ZeroMaxAttempts_RunsOnce(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 0, Delay: 0}
	calls := 0
	_ = p.Execute(context.Background(), func(_ context.Context, _ int) error {
		calls++
		return errFake
	})
	if calls != 1 {
		t.Fatalf("expected 1 call with zero MaxAttempts, got %d", calls)
	}
}
