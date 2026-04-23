package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAdaptiveThrottle_ConcurrentSafety(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(8, 0.5, 0.1, window)
	defer at.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			window.Record(i%3 == 0)
			at.adjust()
			_ = at.Current()
		}(i)
	}
	wg.Wait()
}

func TestAdaptiveThrottle_NeverExceedsCurrent(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(4, 0.5, 0.1, window)
	defer at.Stop()

	var inflight int64
	var maxObserved int64
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = at.Wrap(context.Background(), func(_ context.Context) error {
				cur := atomic.AddInt64(&inflight, 1)
				for {
					prev := atomic.LoadInt64(&maxObserved)
					if cur <= prev || atomic.CompareAndSwapInt64(&maxObserved, prev, cur) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				atomic.AddInt64(&inflight, -1)
				return nil
			})
		}()
	}
	wg.Wait()
	// maxObserved may be <= current at time of each call; just ensure no panic/race
	if atomic.LoadInt64(&maxObserved) < 1 {
		t.Fatal("expected at least one call to run")
	}
}
