package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_ZeroRPS_NoLimit(t *testing.T) {
	rl := NewRateLimiter(0)
	ctx := context.Background()

	// Should return immediately without blocking.
	for i := 0; i < 10; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestRateLimiter_NegativeRPS_NoLimit(t *testing.T) {
	rl := NewRateLimiter(-5)
	ctx := context.Background()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRateLimiter_ContextCancelled_ReturnsError(t *testing.T) {
	// Very low RPS so we block.
	rl := NewRateLimiter(0.01)
	// Drain the initial token.
	_ = rl.Wait(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestRateLimiter_HighRPS_AllowsBurst(t *testing.T) {
	rl := NewRateLimiter(100)
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)
	// 10 calls at 100 RPS should complete well under 200ms.
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected fast burst, took %v", elapsed)
	}
}

func TestRateLimiter_ConcurrentSafety(t *testing.T) {
	rl := NewRateLimiter(500)
	ctx := context.Background()

	var count atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rl.Wait(ctx); err == nil {
				count.Add(1)
			}
		}()
	}

	wg.Wait()
	if count.Load() != 20 {
		t.Errorf("expected 20 successful waits, got %d", count.Load())
	}
}
