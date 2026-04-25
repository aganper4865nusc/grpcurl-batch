package runner

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestCorrelationStore_AssignAndFromContext(t *testing.T) {
	store := NewCorrelationStore(nil)
	ctx, id := store.Assign(context.Background(), "svc", "Method")
	if id == "" {
		t.Fatal("expected non-empty correlation ID")
	}
	if got := FromContext(ctx); got != id {
		t.Fatalf("expected %q from context, got %q", id, got)
	}
}

func TestCorrelationStore_CustomIDFunc(t *testing.T) {
	gen := func(service, method string) string {
		return "custom-" + service + "-" + method
	}
	store := NewCorrelationStore(gen)
	_, id := store.Assign(context.Background(), "payments", "Charge")
	if id != "custom-payments-Charge" {
		t.Fatalf("unexpected ID: %q", id)
	}
}

func TestCorrelationStore_Remove(t *testing.T) {
	store := NewCorrelationStore(nil)
	_, id := store.Assign(context.Background(), "svc", "M")
	if len(store.Snapshot()) != 1 {
		t.Fatal("expected 1 record before remove")
	}
	store.Remove(id)
	if len(store.Snapshot()) != 0 {
		t.Fatal("expected 0 records after remove")
	}
}

func TestCorrelationStore_Snapshot(t *testing.T) {
	store := NewCorrelationStore(nil)
	store.Assign(context.Background(), "svc", "A")
	store.Assign(context.Background(), "svc", "B")
	snap := store.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 records, got %d", len(snap))
	}
}

func TestFromContext_Missing(t *testing.T) {
	if id := FromContext(context.Background()); id != "" {
		t.Fatalf("expected empty ID from plain context, got %q", id)
	}
}

func TestDefaultCorrelationIDFunc_ContainsMethod(t *testing.T) {
	id := DefaultCorrelationIDFunc("svc", "DoThing")
	if !strings.HasPrefix(id, "DoThing-") {
		t.Fatalf("expected ID to start with method name, got %q", id)
	}
}

func TestWithCorrelation_Middleware(t *testing.T) {
	store := NewCorrelationStore(nil)
	var capturedID string
	inner := func(ctx context.Context, call Call) (string, error) {
		capturedID = FromContext(ctx)
		return "ok", nil
	}
	mw := WithCorrelation(store)(inner)
	_, err := mw(context.Background(), Call{Service: "svc", Method: "M"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID == "" {
		t.Fatal("expected correlation ID in context during call")
	}
	// ID should be removed after call completes
	if len(store.Snapshot()) != 0 {
		t.Fatal("expected store to be empty after call")
	}
}

func TestWithCorrelation_ConcurrentSafety(t *testing.T) {
	store := NewCorrelationStore(nil)
	mw := WithCorrelation(store)(func(ctx context.Context, call Call) (string, error) {
		return "ok", nil
	})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mw(context.Background(), Call{Service: "svc", Method: "M"})
		}()
	}
	wg.Wait()
	if len(store.Snapshot()) != 0 {
		t.Fatal("expected empty store after all concurrent calls")
	}
}
