package runner

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWatchdog_ConcurrentCalls_NoDataRace(t *testing.T) {
	w := NewWatchdogPolicy(200 * time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
				PingWatchdog(ctx)
				return "done", nil
			})
		}()
	}
	wg.Wait()
}

func TestWatchdog_StallWhileConcurrent(t *testing.T) {
	w := NewWatchdogPolicy(40 * time.Millisecond)
	results := make([]error, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			_, err := w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(500 * time.Millisecond):
					return "late", nil
				}
			})
			results[idx] = err
		}()
	}
	wg.Wait()
	for i, err := range results {
		if err == nil {
			t.Errorf("call %d: expected stall cancellation, got nil", i)
		}
	}
}
