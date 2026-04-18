package runner

import (
	"context"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

func stageCall(method, address string, tags []string, retries int) manifest.Call {
	return manifest.Call{Method: method, Address: address, Tags: tags, Retries: retries}
}

func TestWithAddressOverride_ReplacesAddress(t *testing.T) {
	stage := WithAddressOverride("newhost:443")
	call, err := stage(context.Background(), stageCall("svc/M", "old:443", nil, 0))
	if err != nil || call.Address != "newhost:443" {
		t.Errorf("expected address override, got %v %v", call.Address, err)
	}
}

func TestWithAddressOverride_EmptyAddress_NoChange(t *testing.T) {
	stage := WithAddressOverride("")
	call, err := stage(context.Background(), stageCall("svc/M", "original:443", nil, 0))
	if err != nil || call.Address != "original:443" {
		t.Errorf("expected original address, got %v", call.Address)
	}
}

func TestWithTagFilter_MatchingTag_Passes(t *testing.T) {
	stage := WithTagFilter("smoke")
	call, err := stage(context.Background(), stageCall("svc/M", "", []string{"smoke", "ci"}, 0))
	if err != nil {
		t.Errorf("expected pass, got error: %v", err)
	}
	_ = call
}

func TestWithTagFilter_NonMatchingTag_Errors(t *testing.T) {
	stage := WithTagFilter("prod")
	_, err := stage(context.Background(), stageCall("svc/M", "", []string{"smoke"}, 0))
	if err == nil {
		t.Error("expected error for non-matching tag")
	}
}

func TestWithMaxRetries_CapsRetries(t *testing.T) {
	stage := WithMaxRetries(2)
	call, err := stage(context.Background(), stageCall("svc/M", "", nil, 5))
	if err != nil || call.Retries != 2 {
		t.Errorf("expected retries=2, got %d %v", call.Retries, err)
	}
}

func TestWithMaxRetries_BelowMax_Unchanged(t *testing.T) {
	stage := WithMaxRetries(5)
	call, err := stage(context.Background(), stageCall("svc/M", "", nil, 2))
	if err != nil || call.Retries != 2 {
		t.Errorf("expected retries=2, got %d", call.Retries)
	}
}

func TestWithDeadlineGuard_NotExceeded_Passes(t *testing.T) {
	p := NewDeadlinePolicy(time.Now().Add(10 * time.Second))
	stage := WithDeadlineGuard(p)
	_, err := stage(context.Background(), stageCall("svc/M", "", nil, 0))
	if err != nil {
		t.Errorf("expected pass, got: %v", err)
	}
}

func TestWithDeadlineGuard_Exceeded_Errors(t *testing.T) {
	p := NewDeadlinePolicy(time.Now().Add(-1 * time.Second))
	stage := WithDeadlineGuard(p)
	_, err := stage(context.Background(), stageCall("svc/M", "", nil, 0))
	if err == nil {
		t.Error("expected error when deadline exceeded")
	}
}

func TestWithTimestampHeader_InjectsHeader(t *testing.T) {
	stage := WithTimestampHeader()
	call, err := stage(context.Background(), stageCall("svc/M", "", nil, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if call.Headers["x-timestamp"] == "" {
		t.Error("expected x-timestamp header to be set")
	}
}
