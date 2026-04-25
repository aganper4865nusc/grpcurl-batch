package runner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalescePolicy_UniqueKeys_BothExecute(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	var calls int64
	next := func(_ context.Context, c Call) (*CallResult, error) {
		atomic.AddInt64(&calls, 1)
		return &CallResult{Call: c}, nil
	}
	wrapped := cp.Wrap(next)

	a := Call{Method: "svc/MethodA"}
	b := Call{Method: "svc/MethodB"}
	wrapped(context.Background(), a) //nolint
	wrapped(context.Background(), b) //nolint

	if atomic.LoadInt64(&calls) != 2 {
		t.Fatalf("expected 2 executions, got %d", calls)
	}
}

func TestCoalescePolicy_SameKey_OnlyOneExecution(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	var calls int64
	ready := make(chan struct{})
	next := func(_ context.Context, c Call) (*CallResult, error) {
		atomic.AddInt64(&calls, 1)
		<-ready
		return &CallResult{Call: c, Response: "shared"}, nil
	}
	wrapped := cp.Wrap(next)

	call := Call{Method: "svc/Method", Address: "localhost:50051"}
	const goroutines = 5
	var wg sync.WaitGroup
	results := make([]*CallResult, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r, _ := wrapped(context.Background(), call)
			results[idx] = r
		}(i)
	}
	time.Sleep(20 * time.Millisecond)
	close(ready)
	wg.Wait()

	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected 1 execution, got %d", calls)
	}
	for _, r := range results {
		if r == nil || r.Response != "shared" {
			t.Fatal("expected all callers to receive the shared result")
		}
	}
}

func TestCoalescePolicy_ErrorPropagatedToAll(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	sentinel := errors.New("boom")
	ready := make(chan struct{})
	next := func(_ context.Context, _ Call) (*CallResult, error) {
		<-ready
		return nil, sentinel
	}
	wrapped := cp.Wrap(next)

	call := Call{Method: "svc/Fail", Address: "localhost:50051"}
	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = wrapped(context.Background(), call)
		}(i)
	}
	time.Sleep(10 * time.Millisecond)
	close(ready)
	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, sentinel) {
			t.Fatalf("caller %d: expected sentinel error, got %v", i, err)
		}
	}
}

func TestCoalescePolicy_Inflight_TracksCorrectly(t *testing.T) {
	cp := NewCoalescePolicy(nil)
	ready := make(chan struct{})
	next := func(_ context.Context, c Call) (*CallResult, error) {
		<-ready
		return &CallResult{Call: c}, nil
	}
	wrapped := cp.Wrap(next)

	call := Call{Method: "svc/X", Address: "localhost:50051"}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped(context.Background(), call) //nolint
	}()
	time.Sleep(10 * time.Millisecond)
	if cp.Inflight() != 1 {
		t.Fatalf("expected 1 inflight, got %d", cp.Inflight())
	}
	close(ready)
	wg.Wait()
	if cp.Inflight() != 0 {
		t.Fatalf("expected 0 inflight after completion, got %d", cp.Inflight())
	}
}
