package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func auditTempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "audit.jsonl")
}

func TestWithAudit_RecordsSuccess(t *testing.T) {
	log, _ := NewAuditLog(auditTempPath(t))
	mw := WithAudit(log)
	call := Call{Service: "svc", Method: "Ping", Address: "localhost:50051"}

	wrapped := mw(func(_ context.Context, c Call) (string, error) {
		return `{"ok":true}`, nil
	})

	res, err := wrapped(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == "" {
		t.Fatal("expected non-empty response")
	}

	events := log.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].Success {
		t.Error("expected event to be successful")
	}
	if events[0].Method != "Ping" {
		t.Errorf("expected method Ping, got %s", events[0].Method)
	}
}

func TestWithAudit_RecordsFailure(t *testing.T) {
	log, _ := NewAuditLog(auditTempPath(t))
	mw := WithAudit(log)
	call := Call{Service: "svc", Method: "Fail", Address: "localhost:50051"}

	wrapped := mw(func(_ context.Context, c Call) (string, error) {
		return "", errors.New("rpc error")
	})

	_, err := wrapped(context.Background(), call)
	if err == nil {
		t.Fatal("expected error")
	}

	events := log.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Success {
		t.Error("expected event to be a failure")
	}
	if events[0].Error != "rpc error" {
		t.Errorf("unexpected error field: %s", events[0].Error)
	}
}

func TestWithAuditFilter_OnlyRecordsMatching(t *testing.T) {
	log, _ := NewAuditLog(auditTempPath(t))
	mw := WithAuditFilter(log, AuditFailuresOnly())
	call := Call{Service: "svc", Method: "Mixed", Address: "localhost:50051"}

	wrapped := mw(func(_ context.Context, c Call) (string, error) {
		return "ok", nil
	})
	_, _ = wrapped(context.Background(), call)

	wrappedFail := mw(func(_ context.Context, c Call) (string, error) {
		return "", errors.New("boom")
	})
	_, _ = wrappedFail(context.Background(), call)

	events := log.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event (failures only), got %d", len(events))
	}
	if events[0].Success {
		t.Error("filtered event should be a failure")
	}
}

func TestWithAudit_Cancelled_StillRecords(t *testing.T) {
	log, _ := NewAuditLog(auditTempPath(t))
	mw := WithAudit(log)
	call := Call{Service: "svc", Method: "Cancel", Address: "localhost:50051"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	wrapped := mw(func(c context.Context, _ Call) (string, error) {
		return "", c.Err()
	})
	_, _ = wrapped(ctx, call)

	if len(log.Events()) != 1 {
		t.Fatal("expected audit record even on cancellation")
	}
}

func TestAuditFailuresOnly_Predicate(t *testing.T) {
	pred := AuditFailuresOnly()

	if pred(AuditEvent{Success: true}) {
		t.Error("should not match successful event")
	}
	if !pred(AuditEvent{Success: false}) {
		t.Error("should match failed event")
	}
}

func init() {
	// ensure os is imported
	_ = os.DevNull
}
