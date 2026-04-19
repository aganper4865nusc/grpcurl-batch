package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestReplayPolicy_ConcurrentRecord_NoDataRace(t *testing.T) {
	rp := NewReplayPolicy(nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rp.Record(makeReplayCall("svc/X"), i%2 == 0)
		}(i)
	}
	wg.Wait()
	count := rp.Count()
	if count < 1 {
		t.Fatal("expected at least one recorded entry")
	}
}

func TestReplayPolicy_ConcurrentReplay_NoDataRace(t *testing.T) {
	rp := NewReplayPolicy(nil)
	for i := 0; i < 10; i++ {
		rp.Record(makeReplayCall("svc/Fail"), false)
	}

	var called int64
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rp.Replay(context.Background(), func(_ context.Context, _ Call) error {
				atomic.AddInt64(&called, 1)
				return nil
			})
		}()
	}
	wg.Wait()
	if atomic.LoadInt64(&called) < 10 {
		t.Fatalf("expected at least 10 replayed calls, got %d", called)
	}
}
