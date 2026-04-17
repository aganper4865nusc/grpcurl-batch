package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

func noopCall(_ context.Context, _ manifest.Call) (string, error) {
	return "ok", nil
}

func failCall(_ context.Context, _ manifest.Call) (string, error) {
	return "", errors.New("fail")
}

func TestChain_SingleMiddleware(t *testing.T) {
	var called bool
	mw := func(next CallFunc) CallFunc {
		return func(ctx context.Context, c manifest.Call) (string, error) {
			called = true
			return next(ctx, c)
		}
	}
	f := Chain(noopCall, mw)
	f(context.Background(), manifest.Call{})
	if !called {
		t.Fatal("middleware was not called")
	}
}

func TestChain_OrderPreserved(t *testing.T) {
	var order []int
	mkMW := func(n int) Middleware {
		return func(next CallFunc) CallFunc {
			return func(ctx context.Context, c manifest.Call) (string, error) {
				order = append(order, n)
				return next(ctx, c)
			}
		}
	}
	f := Chain(noopCall, mkMW(1), mkMW(2), mkMW(3))
	f(context.Background(), manifest.Call{})
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestWithRateLimit_PassesThrough(t *testing.T) {
	rl := NewRateLimiter(0)
	f := Chain(noopCall, WithRateLimit(rl))
	res, err := f(context.Background(), manifest.Call{})
	if err != nil || res != "ok" {
		t.Fatalf("unexpected result: %v %v", res, err)
	}
}

func TestWithCircuitBreaker_PassesThrough(t *testing.T) {
	cb := NewCircuitBreaker(5, time.Second)
	f := Chain(noopCall, WithCircuitBreaker(cb))
	res, err := f(context.Background(), manifest.Call{})
	if err != nil || res != "ok" {
		t.Fatalf("unexpected result: %v %v", res, err)
	}
}

func TestWithBulkhead_PassesThrough(t *testing.T) {
	b := NewBulkheadPolicy(5)
	f := Chain(noopCall, WithBulkhead(b))
	res, err := f(context.Background(), manifest.Call{})
	if err != nil || res != "ok" {
		t.Fatalf("unexpected result: %v %v", res, err)
	}
}

func TestWithBulkhead_RejectsWhenFull(t *testing.T) {
	b := NewBulkheadPolicy(1)
	blocked := make(chan struct{})
	go func() {
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			<-blocked
			return nil
		})
	}()
	time.Sleep(20 * time.Millisecond)
	f := Chain(noopCall, WithBulkhead(b))
	_, err := f(context.Background(), manifest.Call{})
	if !errors.Is(err, ErrBulkheadFull) {
		t.Fatalf("expected ErrBulkheadFull, got %v", err)
	}
	close(blocked)
}
