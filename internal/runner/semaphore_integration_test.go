package runner

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphore_EnforcesMaxConcurrency(t *testing.T) {
	const limit = 3
	const workers = 12
	s := NewSemaphore(limit)

	var active int32
	var maxSeen int32

	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func() {
			_ = s.Acquire(context.Background())
			cur := atomic.AddInt32(&active, 1)
			for {
				prev := atomic.LoadInt32(&maxSeen)
				if cur <= prev || atomic.CompareAndSwapInt32(&maxSeen, prev, cur) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&active, -1)
			s.Release()
			done <- struct{}{}
		}()
	}

	for i := 0; i < workers; i++ {
		<-done
	}

	if maxSeen > int32(limit) {
		t.Errorf("max concurrent exceeded limit: got %d, want <= %d", maxSeen, limit)
	}
}
