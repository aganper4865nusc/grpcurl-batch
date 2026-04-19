package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCooldown_ZeroDuration_NoWait(t *testing.T) {
	cp := NewCooldownPolicy(0)
	cp.RecordFailure()
	start := time.Now()
	if err := cp.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Error("expected no wait for zero duration")
	}
}

func TestCooldown_NoFailure_NoWait(t *testing.T) {
	cp := NewCooldownPolicy(500 * time.Millisecond)
	start := time.Now()
	if err := cp.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Error("expected no wait when no failure recorded")
	}
}

func TestCooldown_WaitsAfterFailure(t *testing.T) {
	cp := NewCooldownPolicy(80 * time.Millisecond)
	cp.RecordFailure()
	start := time.Now()
	if err := cp.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 60*time.Millisecond {
		t.Errorf("expected cooldown wait, got %v", elapsed)
	}
}

func TestCooldown_ContextCancelled_ReturnsErr(t *testing.T) {
	cp := NewCooldownPolicy(2 * time.Second)
	cp.RecordFailure()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := cp.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestCooldown_Wrap_RecordsFailure(t *testing.T) {
	cp := NewCooldownPolicy(100 * time.Millisecond)
	sentinel := errors.New("boom")
	_ = cp.Wrap(context.Background(), func(_ context.Context) error {
		return sentinel
	})
	cp.mu.Lock()
	recorded := !cp.lastFail.IsZero()
	cp.mu.Unlock()
	if !recorded {
		t.Error("expected failure to be recorded after Wrap returns error")
	}
}

func TestCooldown_Wrap_NoRecordOnSuccess(t *testing.T) {
	cp := NewCooldownPolicy(100 * time.Millisecond)
	_ = cp.Wrap(context.Background(), func(_ context.Context) error { return nil })
	cp.mu.Lock()
	recorded := !cp.lastFail.IsZero()
	cp.mu.Unlock()
	if recorded {
		t.Error("expected no failure recorded after successful Wrap")
	}
}
