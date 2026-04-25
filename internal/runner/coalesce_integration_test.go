package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalescePolicy_ConcurrentSafety(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	var executions int64
	next := func(_ context.Context, c Call) (*CallResult, error) {
		atomic.AddInt64(&executions, 1)
		time.Sleep(5 * time.Millisecond)
		return &CallResult{Call: c}, nil
	}
	wrapped := cp.Wrap(next)

	call := Call{Method: "svc/Heavy", Address: "localhost:50051"}
	const workers = 50
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wrapped(context.Background(), call) //nolint
		}()
	}
	wg.Wait()

	// Executions should be far fewer than workers due to coalescing.
	if n := atomic.LoadInt64(&executions); n >= workers {
		t.Fatalf("coalescing failed: %d executions for %d workers", n, workers)
	}
}

func TestCoalescePolicy_SequentialCalls_EachExecutes(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	var executions int64
	next := func(_ context.Context, c Call) (*CallResult, error) {
		atomic.AddInt64(&executions, 1)
		return &CallResult{Call: c}, nil
	}
	wrapped := cp.Wrap(next)

	call := Call{Method: "svc/Seq", Address: "localhost:50051"}
	for i := 0; i < 5; i++ {
		wrapped(context.Background(), call) //nolint
	}

	if n := atomic.LoadInt64(&executions); n != 5 {
		t.Fatalf("expected 5 sequential executions, got %d", n)
	}
}
