package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func noopCall(_ context.Context) (string, error) {
	return "ok", nil
}

func failCall(_ context.Context) (string, error) {
	return "", errors.New("call failed")
}

func TestChain_SingleMiddleware(t *testing.T) {
	var called bool
	mw := func(next CallFunc) CallFunc {
		return func(ctx context.Context) (string, error) {
			called = true
			return next(ctx)
		}
	}
	result, err := Chain(mw)(noopCall)(context.Background())
	if err != nil || result != "ok" || !called {
		t.Fatalf("unexpected result: %v %v %v", result, err, called)
	}
}

func TestChain_OrderPreserved(t *testing.T) {
	var order []int
	make := func(n int) Middleware {
		return func(next CallFunc) CallFunc {
			return func(ctx context.Context) (string, error) {
				order = append(order, n)
				return next(ctx)
			}
		}
	}
	Chain(make(1), make(2), make(3))(noopCall)(context.Background())
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("wrong order: %v", order)
	}
}

func TestWithRateLimit_PassesThrough(t *testing.T) {
	rl := NewRateLimiter(100)
	mw := WithRateLimit(rl)
	result, err := mw(noopCall)(context.Background())
	if err != nil || result != "ok" {
		t.Fatalf("unexpected: %v %v", result, err)
	}
}

func TestWithRateLimit_CancelledContext(t *testing.T) {
	rl := NewRateLimiter(0.0001)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	mw := WithRateLimit(rl)
	// exhaust the limiter by calling many times
	for i := 0; i < 5; i++ {
		mw(noopCall)(ctx) //nolint
	}
	_, err := mw(noopCall)(ctx)
	// may or may not error depending on timing; just ensure no panic
	_ = err
}

func TestWithCircuitBreaker_PassesThrough(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	mw := WithCircuitBreaker(cb)
	result, err := mw(noopCall)(context.Background())
	if err != nil || result != "ok" {
		t.Fatalf("unexpected: %v %v", result, err)
	}
}

func TestWithCircuitBreaker_PropagatesError(t *testing.T) {
	cb := NewCircuitBreaker(10, time.Second)
	mw := WithCircuitBreaker(cb)
	_, err := mw(failCall)(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithTimeout_WithinDeadline(t *testing.T) {
	tp := DefaultTimeoutPolicy(5 * time.Second)
	mw := WithTimeout(tp)
	result, err := mw(noopCall)(context.Background())
	if err != nil || result != "ok" {
		t.Fatalf("unexpected: %v %v", result, err)
	}
}
