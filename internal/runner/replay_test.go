package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func makeReplayCall(method string) Call {
	return Call{Method: method, Address: "localhost:50051"}
}

func TestReplayPolicy_RecordFailedOnly(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/Fail"), false)
	rp.Record(makeReplayCall("svc/OK"), true)
	if rp.Count() != 1 {
		t.Fatalf("expected 1 entry, got %d", rp.Count())
	}
	if rp.Entries()[0].Call.Method != "svc/Fail" {
		t.Error("expected failed call recorded")
	}
}

func TestReplayPolicy_CustomFilter(t *testing.T) {
	rp := NewReplayPolicy(func(e ReplayEntry) bool { return true })
	rp.Record(makeReplayCall("svc/OK"), true)
	if rp.Count() != 1 {
		t.Fatalf("expected 1, got %d", rp.Count())
	}
}

func TestReplayPolicy_Reset(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/Fail"), false)
	rp.Reset()
	if rp.Count() != 0 {
		t.Fatal("expected empty after reset")
	}
}

func TestReplayPolicy_Replay_Success(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/A"), false)
	rp.Record(makeReplayCall("svc/B"), false)

	replayed := []string{}
	var mu sync.Mutex
	errs := rp.Replay(context.Background(), func(_ context.Context, c Call) error {
		mu.Lock()
		replayed = append(replayed, c.Method)
		mu.Unlock()
		return nil
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(replayed) != 2 {
		t.Fatalf("expected 2 replayed, got %d", len(replayed))
	}
}

func TestReplayPolicy_Replay_CollectsErrors(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/Fail"), false)

	errs := rp.Replay(context.Background(), func(_ context.Context, _ Call) error {
		return errors.New("still failing")
	})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestReplayPolicy_Replay_ContextCancelled(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/A"), false)
	rp.Record(makeReplayCall("svc/B"), false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errs := rp.Replay(ctx, func(_ context.Context, _ Call) error { return nil })
	if len(errs) == 0 {
		t.Fatal("expected cancellation error")
	}
}

func TestReplayPolicy_Entries_IsSnapshot(t *testing.T) {
	rp := NewReplayPolicy(nil)
	rp.Record(makeReplayCall("svc/A"), false)
	snap := rp.Entries()
	rp.Reset()
	if len(snap) != 1 {
		t.Fatal("snapshot should not be affected by reset")
	}
}
