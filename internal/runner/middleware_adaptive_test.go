package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func okAdaptiveCall(_ context.Context, _ Call) (Response, error) {
	return Response{Body: "ok"}, nil
}

func errAdaptiveCall(_ context.Context, _ Call) (Response, error) {
	return Response{}, errors.New("fail")
}

func TestWithAdaptiveThrottle_PassesThroughSuccess(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(4, 0.5, 0.1, window)
	defer at.Stop()
	mw := WithAdaptiveThrottle(at, window)
	call := Call{Method: "/pkg.Svc/Method"}
	resp, err := mw(okAdaptiveCall)(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Body != "ok" {
		t.Fatalf("unexpected body: %s", resp.Body)
	}
}

func TestWithAdaptiveThrottle_RecordsErrorsInWindow(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(4, 0.5, 0.1, window)
	defer at.Stop()
	mw := WithAdaptiveThrottle(at, window)
	call := Call{Method: "/pkg.Svc/Method"}
	for i := 0; i < 5; i++ {
		_, _ = mw(errAdaptiveCall)(context.Background(), call)
	}
	stats := window.Stats()
	if stats.Failures != 5 {
		t.Fatalf("expected 5 failures recorded, got %d", stats.Failures)
	}
}

func TestWithAdaptiveThrottle_CancelledContext(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(1, 0.5, 0.1, window)
	defer at.Stop()
	// Block the single slot
	blockCh := make(chan struct{})
	go func() {
		_, _ = mw_block(at, window, blockCh)(context.Background(), Call{})
	}()
	time.Sleep(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mw := WithAdaptiveThrottle(at, window)
	_, err := mw(okAdaptiveCall)(ctx, Call{})
	close(blockCh)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func mw_block(at *AdaptiveThrottle, window *WindowPolicy, ch chan struct{}) CallFunc {
	mw := WithAdaptiveThrottle(at, window)
	return mw(func(_ context.Context, _ Call) (Response, error) {
		<-ch
		return Response{}, nil
	})
}

func TestAdaptiveThrottleStatus_ContainsFields(t *testing.T) {
	window := NewWindowPolicy(10 * time.Second)
	at := NewAdaptiveThrottle(8, 0.5, 0.1, window)
	defer at.Stop()
	window.Record(true)
	window.Record(false)
	s := AdaptiveThrottleStatus(at, window)
	if s == "" {
		t.Fatal("expected non-empty status string")
	}
	for _, want := range []string{"current=", "err_rate=", "total="} {
		if !contains(s, want) {
			t.Errorf("status missing %q: %s", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
