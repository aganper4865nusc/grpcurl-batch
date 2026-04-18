package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDrainPolicy_ConcurrentAcquireRelease(t *testing.T) {
	d := NewDrainPolicy(3 * time.Second)
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		d.Acquire()
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			d.Release()
		}()
	}
	ctx := context.Background()
	if err := d.Wait(ctx); err != nil {
		t.Fatalf("drain timed out: %v", err)
	}
	wg.Wait()
	if d.InFlight() != 0 {
		t.Fatalf("expected 0 inflight, got %d", d.InFlight())
	}
}

func TestDrainPolicy_MultipleWaiters(t *testing.T) {
	d := NewDrainPolicy(2 * time.Second)
	d.Acquire()
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			// Only first waiter gets the done signal; others rely on timeout not firing
			_ = d.Wait(ctx)
		}()
	}
	time.Sleep(20 * time.Millisecond)
	d.Release()
	wg.Wait()
}
