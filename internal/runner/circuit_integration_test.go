package runner

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCircuitBreaker_ConcurrentSafety verifies the circuit breaker is safe
// under concurrent access from multiple goroutines.
func TestCircuitBreaker_ConcurrentSafety(t *testing.T) {
	cb := NewCircuitBreaker(10, time.Second)
	var wg sync.WaitGroup
	const workers = 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
			}
			_ = cb.Allow()
			_ = cb.State()
		}(i)
	}
	wg.Wait()
	// No race condition or panic — test passes if we reach here.
}

// TestCircuitBreaker_BlocksCallsWhenOpen ensures that once the circuit opens,
// all subsequent Allow calls return ErrCircuitOpen until the timeout elapses.
func TestCircuitBreaker_BlocksCallsWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 200*time.Millisecond)
	cb.RecordFailure()
	cb.RecordFailure()

	var blocked int32
	var wg sync.WaitGroup
	const callers = 20

	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cb.Allow(); err == ErrCircuitOpen {
				atomic.AddInt32(&blocked, 1)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&blocked) != callers {
		t.Fatalf("expected all %d callers blocked, got %d", callers, blocked)
	}

	// After timeout, circuit should allow a probe.
	time.Sleep(210 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected allow after reset timeout, got %v", err)
	}
}
