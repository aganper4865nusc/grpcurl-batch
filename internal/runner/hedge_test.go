package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestHedgePolicy_ZeroMaxHedge_NoExtraCalls(t *testing.T) {
	calls := int32(0)
	h := NewHedgePolicy(10*time.Millisecond, 0)
	resp, err := h.Execute(context.Background(), func(ctx context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %s", resp)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestHedgePolicy_FirstCallSucceeds_NoHedge(t *testing.T) {
	calls := int32(0)
	h := NewHedgePolicy(50*time.Millisecond, 1)
	resp, err := h.Execute(context.Background(), func(ctx context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "fast", nil
	})
	if err != nil || resp != "fast" {
		t.Fatalf("unexpected: %v %v", resp, err)
	}
}

func TestHedgePolicy_SlowFirst_HedgeWins(t *testing.T) {
	calls := int32(0)
	h := NewHedgePolicy(20*time.Millisecond, 1)
	resp, err := h.Execute(context.Background(), func(ctx context.Context) (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			time.Sleep(200 * time.Millisecond)
		}
		return "hedge", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hedge" {
		t.Fatalf("expected hedge, got %s", resp)
	}
}

func TestHedgePolicy_AllFail_ReturnsError(t *testing.T) {
	h := NewHedgePolicy(10*time.Millisecond, 1)
	err := errors.New("fail")
	_, got := h.Execute(context.Background(), func(ctx context.Context) (string, error) {
		return "", err
	})
	if got == nil {
		t.Fatal("expected error")
	}
}

func TestHedgePolicy_ContextCancelled_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := NewHedgePolicy(10*time.Millisecond, 1)
	_, err := h.Execute(ctx, func(ctx context.Context) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "ok", nil
	})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestHedgePolicy_NegativeMaxHedge_Clamped(t *testing.T) {
	h := NewHedgePolicy(10*time.Millisecond, -5)
	if h.maxHedge != 0 {
		t.Fatalf("expected 0, got %d", h.maxHedge)
	}
}
