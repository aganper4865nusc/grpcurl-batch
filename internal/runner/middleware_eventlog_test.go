package runner

import (
	"context"
	"errors"
	"testing"
)

func okEventCall(_ context.Context, _ Call) (string, error) {
	return "ok", nil
}

func errEventCall(_ context.Context, _ Call) (string, error) {
	return "", errors.New("boom")
}

func TestWithEventLogMiddleware_RecordsSuccess(t *testing.T) {
	el := NewEventLog(20)
	ctx := WithEventLog(context.Background(), el)

	mw := WithEventLogMiddleware(okEventCall)
	_, err := mw(ctx, Call{ID: "c1", Method: "Ping"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(el.Filter(EventCallStarted)); got != 1 {
		t.Errorf("expected 1 started event, got %d", got)
	}
	if got := len(el.Filter(EventCallSucceeded)); got != 1 {
		t.Errorf("expected 1 succeeded event, got %d", got)
	}
	if got := len(el.Filter(EventCallFailed)); got != 0 {
		t.Errorf("expected 0 failed events, got %d", got)
	}
}

func TestWithEventLogMiddleware_RecordsFailure(t *testing.T) {
	el := NewEventLog(20)
	ctx := WithEventLog(context.Background(), el)

	mw := WithEventLogMiddleware(errEventCall)
	_, err := mw(ctx, Call{ID: "c2", Method: "Ping"})
	if err == nil {
		t.Fatal("expected error")
	}

	if got := len(el.Filter(EventCallFailed)); got != 1 {
		t.Errorf("expected 1 failed event, got %d", got)
	}
	events := el.Filter(EventCallFailed)
	if events[0].Err == "" {
		t.Error("expected error message in event")
	}
}

func TestWithEventLogMiddleware_NoLog_IsNoop(t *testing.T) {
	// No EventLog in context — should not panic.
	mw := WithEventLogMiddleware(okEventCall)
	_, err := mw(context.Background(), Call{ID: "c3", Method: "Ping"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEventLogSummary_Counts(t *testing.T) {
	el := NewEventLog(50)
	ctx := WithEventLog(context.Background(), el)

	mwOK := WithEventLogMiddleware(okEventCall)
	mwErr := WithEventLogMiddleware(errEventCall)

	for i := 0; i < 3; i++ {
		_, _ = mwOK(ctx, Call{ID: "ok", Method: "M"})
	}
	for i := 0; i < 2; i++ {
		_, _ = mwErr(ctx, Call{ID: "err", Method: "M"})
	}

	summary := EventLogSummary(el)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	// Smoke-check that counts appear in the string.
	for _, want := range []string{"started=5", "succeeded=3", "failed=2"} {
		if !containsStr(summary, want) {
			t.Errorf("summary %q missing %q", summary, want)
		}
	}
}

func TestEventLogSummary_NilLog(t *testing.T) {
	got := EventLogSummary(nil)
	if got != "no event log" {
		t.Errorf("unexpected: %q", got)
	}
}

// containsStr is a small helper to avoid importing strings in tests.
func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
