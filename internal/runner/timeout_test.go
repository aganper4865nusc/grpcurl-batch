package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeoutPolicy_SuccessWithinDeadline(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 500 * time.Millisecond}
	err := tp.Apply(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestTimeoutPolicy_ExceedsDeadline(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 50 * time.Millisecond}
	err := tp.Apply(context.Background(), func(ctx context.Context) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutPolicy_ZeroTimeout_NoDeadline(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 0}
	called := false
	err := tp.Apply(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected fn to be called")
	}
}

func TestTimeoutPolicy_PropagatesError(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 500 * time.Millisecond}
	sentinel := errors.New("sentinel error")
	err := tp.Apply(context.Background(), func(_ context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutPolicy_CancelledParentContext(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 500 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := tp.Apply(ctx, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got %v", err)
	}
}

func TestDefaultTimeoutPolicy_Is30Seconds(t *testing.T) {
	tp := DefaultTimeoutPolicy()
	if tp.Timeout != 30*time.Second {
		t.Fatalf("expected 30s, got %v", tp.Timeout)
	}
}

func TestTimeoutPolicy_ContextPassedToFn(t *testing.T) {
	tp := TimeoutPolicy{Timeout: 500 * time.Millisecond}
	var deadline time.Time
	var ok bool
	_ = tp.Apply(context.Background(), func(ctx context.Context) error {
		deadline, ok = ctx.Deadline()
		return nil
	})
	if !ok {
		t.Fatal("expected context to have a deadline set")
	}
	if time.Until(deadline) > 500*time.Millisecond {
		t.Fatalf("deadline too far in the future: %v", deadline)
	}
}
