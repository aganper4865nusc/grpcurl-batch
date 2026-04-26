package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func shadowCall(method string) Call {
	return Call{Method: method, Address: "localhost:50051"}
}

func TestShadowPolicy_PrimarySucceeds_ShadowFired(t *testing.T) {
	var fired int32
	s := NewShadowPolicy(func(_ context.Context, _ Call) error {
		atomic.AddInt32(&fired, 1)
		return nil
	}, 0)

	wrapped := s.Wrap(func(_ context.Context, _ Call) error { return nil })
	if err := wrapped(context.Background(), shadowCall("/svc/Method")); err != nil {
		t.Fatalf("unexpected primary error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&fired) != 1 {
		t.Errorf("expected shadow to fire once, got %d", fired)
	}
}

func TestShadowPolicy_PrimaryFails_ShadowStillFired(t *testing.T) {
	var fired int32
	s := NewShadowPolicy(func(_ context.Context, _ Call) error {
		atomic.AddInt32(&fired, 1)
		return nil
	}, 10)

	primaryErr := errors.New("primary error")
	wrapped := s.Wrap(func(_ context.Context, _ Call) error { return primaryErr })
	err := wrapped(context.Background(), shadowCall("/svc/Fail"))
	if !errors.Is(err, primaryErr) {
		t.Fatalf("expected primary error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&fired) != 1 {
		t.Errorf("shadow should fire even when primary fails")
	}
}

func TestShadowPolicy_ResultsRecorded(t *testing.T) {
	s := NewShadowPolicy(func(_ context.Context, _ Call) error { return nil }, 10)
	wrapped := s.Wrap(func(_ context.Context, _ Call) error { return nil })
	_ = wrapped(context.Background(), shadowCall("/svc/A"))
	time.Sleep(60 * time.Millisecond)

	results := s.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 shadow result, got %d", len(results))
	}
	if results[0].Call.Method != "/svc/A" {
		t.Errorf("unexpected method: %s", results[0].Call.Method)
	}
}

func TestShadowPolicy_EvictsOldestWhenFull(t *testing.T) {
	s := NewShadowPolicy(func(_ context.Context, _ Call) error { return nil }, 2)
	wrapped := s.Wrap(func(_ context.Context, _ Call) error { return nil })

	for _, m := range []string{"/a", "/b", "/c"} {
		_ = wrapped(context.Background(), shadowCall(m))
	}
	time.Sleep(80 * time.Millisecond)

	results := s.Results()
	if len(results) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(results))
	}
}

func TestShadowPolicy_Reset_ClearsResults(t *testing.T) {
	s := NewShadowPolicy(func(_ context.Context, _ Call) error { return nil }, 10)
	wrapped := s.Wrap(func(_ context.Context, _ Call) error { return nil })
	_ = wrapped(context.Background(), shadowCall("/svc/X"))
	time.Sleep(50 * time.Millisecond)

	s.Reset()
	if got := s.Results(); len(got) != 0 {
		t.Errorf("expected empty results after reset, got %d", len(got))
	}
}
