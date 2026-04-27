package runner

import (
	"context"
	"errors"
	"testing"
)

func probeOK(_ context.Context, _ Call) error { return nil }
func probeFail(_ context.Context, _ Call) error { return errors.New("unhealthy") }

func TestWithProbe_HealthyTarget_ForwardsCall(t *testing.T) {
	p := NewProbePolicy("svc:50051", probeOK, 3)
	mw := WithProbe(p)

	called := false
	next := func(_ context.Context, _ Call) (string, error) {
		called = true
		return "ok", nil
	}

	out, err := mw(next)(context.Background(), Call{Method: "Ping"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected 'ok', got %q", out)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
}

func TestWithProbe_UnhealthyTarget_BlocksCall(t *testing.T) {
	p := NewProbePolicy("svc:50051", probeFail, 1)
	// exhaust the threshold so the probe is marked unhealthy
	_ = p.Check(context.Background(), Call{})

	mw := WithProbe(p)
	next := func(_ context.Context, _ Call) (string, error) {
		t.Fatal("next should not be called")
		return "", nil
	}

	_, err := mw(next)(context.Background(), Call{Method: "Ping"})
	if err == nil {
		t.Fatal("expected error from probe gate")
	}
}

func TestProbeSnapshot_ReflectsState(t *testing.T) {
	p := NewProbePolicy("svc:50051", probeOK, 3)
	snap := ProbeSnapshot(p)

	if snap.Target != "svc:50051" {
		t.Errorf("expected target 'svc:50051', got %q", snap.Target)
	}
	if !snap.Healthy {
		t.Error("expected healthy=true on fresh policy")
	}
	if snap.FailCount != 0 {
		t.Errorf("expected failCount=0, got %d", snap.FailCount)
	}
}

func TestWithProbe_CancelledContext_ReturnsCtxErr(t *testing.T) {
	p := NewProbePolicy("svc:50051", probeOK, 3)
	mw := WithProbe(p)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	next := func(ctx context.Context, _ Call) (string, error) {
		return "", ctx.Err()
	}

	_, err := mw(next)(ctx, Call{Method: "Ping"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
