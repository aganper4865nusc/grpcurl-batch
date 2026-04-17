package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestThrottlePolicy_ZeroMax_NoLimit(t *testing.T) {
	th := NewThrottlePolicy(0, time.Second)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		if err := th.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestThrottlePolicy_AllowsUpToMax(t *testing.T) {
	th := NewThrottlePolicy(3, time.Second)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := th.Wait(ctx); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
	if got := th.Count(); got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
}

func TestThrottlePolicy_BlocksWhenFull(t *testing.T) {
	th := NewThrottlePolicy(2, 200*time.Millisecond)
	ctx := context.Background()
	_ = th.Wait(ctx)
	_ = th.Wait(ctx)

	start := time.Now()
	_ = th.Wait(ctx)
	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Fatalf("expected throttle delay, got %v", elapsed)
	}
}

func TestThrottlePolicy_ContextCancelled(t *testing.T) {
	th := NewThrottlePolicy(1, 10*time.Second)
	ctx := context.Background()
	_ = th.Wait(ctx)

	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := th.Wait(ctx2)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestThrottlePolicy_ConcurrentSafety(t *testing.T) {
	th := NewThrottlePolicy(50, time.Second)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = th.Wait(ctx)
		}()
	}
	wg.Wait()
	if got := th.Count(); got != 50 {
		t.Fatalf("expected 50, got %d", got)
	}
}

func TestThrottlePolicy_ResetsAfterWindow(t *testing.T) {
	th := NewThrottlePolicy(2, 100*time.Millisecond)
	ctx := context.Background()
	_ = th.Wait(ctx)
	_ = th.Wait(ctx)

	// Wait for the window to expire, then confirm we can proceed without blocking.
	time.Sleep(150 * time.Millisecond)
	start := time.Now()
	_ = th.Wait(ctx)
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("expected no throttle delay after window reset, got %v", elapsed)
	}
}
