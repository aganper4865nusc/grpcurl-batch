package runner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrencyLimiter_ZeroMax_DefaultsToOne(t *testing.T) {
	cl := NewConcurrencyLimiter(0)
	if cl.Max() != 1 {
		t.Fatalf("expected max=1, got %d", cl.Max())
	}
}

func TestConcurrencyLimiter_AcquireRelease(t *testing.T) {
	cl := NewConcurrencyLimiter(3)
	ctx := context.Background()

	if err := cl.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.Inflight() != 1 {
		t.Fatalf("expected 1 inflight, got %d", cl.Inflight())
	}
	cl.Release()
	if cl.Inflight() != 0 {
		t.Fatalf("expected 0 inflight after release, got %d", cl.Inflight())
	}
}

func TestConcurrencyLimiter_BlocksWhenFull(t *testing.T) {
	cl := NewConcurrencyLimiter(2)
	ctx := context.Background()

	if err := cl.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cl.Acquire(ctx); err != nil {
		t.Fatal(err)
	}

	err := cl.Acquire(ctx)
	if !errors.Is(err, ErrConcurrencyLimitExceeded) {
		t.Fatalf("expected ErrConcurrencyLimitExceeded, got %v", err)
	}
	// inflight must remain at 2 after the rejected acquire
	if cl.Inflight() != 2 {
		t.Fatalf("expected inflight=2, got %d", cl.Inflight())
	}
}

func TestConcurrencyLimiter_CancelledContext(t *testing.T) {
	cl := NewConcurrencyLimiter(5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cl.Acquire(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestConcurrencyLimiter_Wrap_Success(t *testing.T) {
	cl := NewConcurrencyLimiter(2)
	ctx := context.Background()

	err := cl.Wrap(ctx, func(_ context.Context) error {
		if cl.Inflight() != 1 {
			t.Errorf("expected 1 inflight inside Wrap, got %d", cl.Inflight())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.Inflight() != 0 {
		t.Fatalf("expected 0 inflight after Wrap, got %d", cl.Inflight())
	}
}

func TestConcurrencyLimiter_ConcurrentSafety(t *testing.T) {
	const max = 10
	cl := NewConcurrencyLimiter(max)
	ctx := context.Background()

	var peak atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cl.Wrap(ctx, func(_ context.Context) error {
				current := cl.Inflight()
				for {
					old := peak.Load()
					if current <= old || peak.CompareAndSwap(old, current) {
						break
					}
				}
				time.Sleep(time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()

	if p := peak.Load(); p > int64(max) {
		t.Fatalf("peak inflight %d exceeded max %d", p, max)
	}
}
