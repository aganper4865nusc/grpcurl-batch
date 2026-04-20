package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func successProbe(_ context.Context, _ string) error { return nil }
func failProbe(_ context.Context, _ string) error    { return errors.New("unreachable") }

func TestProbePolicy_HealthyTarget_AllowsCall(t *testing.T) {
	p := NewProbePolicy(successProbe, time.Second, 3)
	out, err := p.Wrap(context.Background(), "localhost:50051", func(_ context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected 'ok', got %q", out)
	}
}

func TestProbePolicy_UnhealthyTarget_BlocksCall(t *testing.T) {
	p := NewProbePolicy(failProbe, time.Second, 3)
	called := false
	_, err := p.Wrap(context.Background(), "localhost:50051", func(_ context.Context) (string, error) {
		called = true
		return "", nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if called {
		t.Fatal("fn should not have been called")
	}
}

func TestProbePolicy_MaxFails_ShortCircuits(t *testing.T) {
	p := NewProbePolicy(failProbe, time.Second, 2)
	ctx := context.Background()
	// exhaust failures via Check
	p.Check(ctx, "addr")
	p.Check(ctx, "addr")
	_, err := p.Wrap(ctx, "addr", func(_ context.Context) (string, error) {
		return "", nil
	})
	if err == nil {
		t.Fatal("expected short-circuit error")
	}
}

func TestProbePolicy_RecoveryResetsFails(t *testing.T) {
	p := NewProbePolicy(successProbe, time.Second, 3)
	p.failCount.Store(2)
	res := p.Check(context.Background(), "addr")
	if !res.Healthy {
		t.Fatal("expected healthy")
	}
	if p.failCount.Load() != 0 {
		t.Fatalf("expected failCount reset to 0, got %d", p.failCount.Load())
	}
}

func TestProbePolicy_LastResult_NilBeforeFirstCheck(t *testing.T) {
	p := NewProbePolicy(successProbe, time.Second, 3)
	if p.LastResult() != nil {
		t.Fatal("expected nil before any check")
	}
}

func TestProbePolicy_Check_RecordsLatency(t *testing.T) {
	p := NewProbePolicy(successProbe, time.Second, 3)
	res := p.Check(context.Background(), "addr")
	if res.Latency < 0 {
		t.Fatal("latency should be non-negative")
	}
	if p.LastResult() == nil {
		t.Fatal("LastResult should be set after Check")
	}
}
