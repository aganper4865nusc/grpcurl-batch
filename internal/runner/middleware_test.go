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
	called := false
	mw := func(next CallFunc) CallFunc {
		return func(ctx context.Context, c manifest.Call) (string, error) {
			called = true
			return next(ctx, c)
		}
	}
	f := Chain(noopCall, mw)
	f(context.Background(), manifest.Call{})
	if !called {
		t.Fatal("middleware not called")
	}
}

func TestChain_OrderPreserved(t *testing.T) {
	var order []int
	make := func(n int) Middleware {
		return func(next CallFunc) CallFunc {
			return func(ctx context.Context, c manifest.Call) (string, error) {
				order = append(order, n)
				return next(ctx, c)
			}
		}
	}
	f := Chain(noopCall, make(1), make(2), make(3))
	f(context.Background(), manifest.Call{})
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("unexpected order %v", order)
	}
}

func TestWithRateLimit_PassesThrough(t *testing.T) {
	rl := NewRateLimiter(0)
	f := Chain(noopCall, WithRateLimit(rl))
	out, err := f(context.Background(), manifest.Call{})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected result %q %v", out, err)
	}
}

func TestWithTimeout_PassesThrough(t *testing.T) {
	tp := DefaultTimeoutPolicy(5 * time.Second)
	f := Chain(noopCall, WithTimeout(tp))
	out, err := f(context.Background(), manifest.Call{})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected result %q %v", out, err)
	}
}

func TestWithDrain_TracksInflight(t *testing.T) {
	d := NewDrainPolicy(time.Second)
	var inflight int
	tracking := func(_ context.Context, _ manifest.Call) (string, error) {
		inflight = d.InFlight()
		return "ok", nil
	}
	f := Chain(tracking, WithDrain(d))
	f(context.Background(), manifest.Call{})
	if inflight != 1 {
		t.Fatalf("expected 1 inflight during call, got %d", inflight)
	}
	if d.InFlight() != 0 {
		t.Fatalf("expected 0 after call, got %d", d.InFlight())
	}
}

func TestWithBulkhead_PassesThrough(t *testing.T) {
	b := NewBulkheadPolicy(2)
	f := Chain(noopCall, WithBulkhead(b))
	out, err := f(context.Background(), manifest.Call{})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected %q %v", out, err)
	}
}
