package runner_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/grpcurl-batch/internal/manifest"
	"github.com/your-org/grpcurl-batch/internal/runner"
)

type mockExecutor struct {
	callCount atomic.Int32
	failTimes int
	output    string
}

func (m *mockExecutor) Execute(_ context.Context, _ manifest.Call) (string, error) {
	count := int(m.callCount.Add(1))
	if count <= m.failTimes {
		return "", errors.New("transient error")
	}
	return m.output, nil
}

func TestRunner_SuccessOnFirstAttempt(t *testing.T) {
	exec := &mockExecutor{output: `{"status":"ok"}`}
	r := runner.New(exec, 2)
	calls := []manifest.Call{{Name: "test-call", Retries: 2}}

	results := r.Run(context.Background(), calls)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
	if results[0].Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", results[0].Attempts)
	}
}

func TestRunner_SuccessAfterRetry(t *testing.T) {
	exec := &mockExecutor{failTimes: 2, output: "pong"}
	r := runner.New(exec, 1)
	calls := []manifest.Call{{Name: "retry-call", Retries: 3, RetryDelay: time.Millisecond}}

	results := r.Run(context.Background(), calls)

	if results[0].Err != nil {
		t.Errorf("expected success after retries, got: %v", results[0].Err)
	}
	if results[0].Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", results[0].Attempts)
	}
}

func TestRunner_ExhaustsRetries(t *testing.T) {
	exec := &mockExecutor{failTimes: 10}
	r := runner.New(exec, 1)
	calls := []manifest.Call{{Name: "fail-call", Retries: 2, RetryDelay: time.Millisecond}}

	results := r.Run(context.Background(), calls)

	if results[0].Err == nil {
		t.Error("expected error after exhausting retries")
	}
	if results[0].Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", results[0].Attempts)
	}
}

func TestRunner_ConcurrencyLimit(t *testing.T) {
	exec := &mockExecutor{output: "ok"}
	r := runner.New(exec, 2)

	calls := make([]manifest.Call, 6)
	for i := range calls {
		calls[i] = manifest.Call{Name: "call"}
	}

	results := r.Run(context.Background(), calls)
	if len(results) != 6 {
		t.Errorf("expected 6 results, got %d", len(results))
	}
}
