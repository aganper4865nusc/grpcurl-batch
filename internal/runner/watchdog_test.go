package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWatchdog_SuccessWithinStall(t *testing.T) {
	w := NewWatchdogPolicy(500 * time.Millisecond)
	res, err := w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "ok" {
		t.Fatalf("expected ok, got %s", res)
	}
}

func TestWatchdog_StallCancelsContext(t *testing.T) {
	w := NewWatchdogPolicy(50 * time.Millisecond)
	_, err := w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(2 * time.Second):
			return "late", nil
		}
	})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestWatchdog_PingResetsTimer(t *testing.T) {
	w := NewWatchdogPolicy(80 * time.Millisecond)
	res, err := w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
		// ping three times, each within the stall window
		for i := 0; i < 3; i++ {
			time.Sleep(40 * time.Millisecond)
			PingWatchdog(ctx)
		}
		return "pong", nil
	})
	if err != nil {
		t.Fatalf("unexpected error after pings: %v", err)
	}
	if res != "pong" {
		t.Fatalf("expected pong, got %s", res)
	}
}

func TestWatchdog_PropagatesError(t *testing.T) {
	w := NewWatchdogPolicy(500 * time.Millisecond)
	sentinel := errors.New("call failed")
	_, err := w.Wrap(context.Background(), func(ctx context.Context) (string, error) {
		return "", sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWatchdog_ZeroTimeout_DefaultsTo10s(t *testing.T) {
	w := NewWatchdogPolicy(0)
	if w.stallTimeout != 10*time.Second {
		t.Fatalf("expected 10s default, got %v", w.stallTimeout)
	}
}

func TestPingWatchdog_NoWatchdog_NoOp(t *testing.T) {
	// Should not panic when no watchdog is in context.
	PingWatchdog(context.Background())
}
