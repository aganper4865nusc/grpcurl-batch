package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBulkhead_MaxConcurrencyNeverExceeded(t *testing.T) {
	const max = 4
	b := NewBulkheadPolicy(max)
	var peak atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Do(context.Background(), func(ctx context.Context) error {
				current := b.Active()
				for {
					old := peak.Load()
					if current <= old || peak.CompareAndSwap(old, current) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()
	if peak.Load() > max {
		t.Fatalf("peak concurrency %d exceeded max %d", peak.Load(), max)
	}
}

func TestBulkhead_RejectedCallsDoNotBlock(t *testing.T) {
	b := NewBulkheadPolicy(1)
	blocked := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			<-blocked
			return nil
		})
	}()
	time.Sleep(10 * time.Millisecond)
	start := time.Now()
	for i := 0; i < 10; i++ {
		_ = b.Do(context.Background(), func(ctx context.Context) error { return nil })
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("rejected calls should return immediately")
	}
	close(blocked)
	wg.Wait()
}
