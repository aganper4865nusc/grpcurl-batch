package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

type mockExecutor struct {
	callCount atomic.Int32
	failTimes int
	output    string
}

func (m *mockExecutor) Execute(_ context.Context, _ manifest.Call) (string, error) {
	n := int(m.callCount.Add(1))
	if n <= m.failTimes {
		return "", errors.New("mock error")
	}
	return m.output, nil
}

func makeCall(name string, retries int) manifest.Call {
	return manifest.Call{
		Name:    name,
		Address: "localhost:50051",
		Method:  "pkg.Service/Method",
		Retries: retries,
	}
}

func TestRunner_SuccessOnFirstAttempt(t *testing.T) {
	exec := &mockExecutor{output: `{"ok":true}`}
	r := New(exec, 1)
	m := &manifest.Manifest{Calls: []manifest.Call{makeCall("c1", 0)}}
	s := r.Run(context.Background(), m)
	if s.Succeeded != 1 || s.Failed != 0 {
		t.Fatalf("expected 1 success, got %+v", s)
	}
}

func TestRunner_SuccessAfterRetry(t *testing.T) {
	exec := &mockExecutor{failTimes: 1, output: "ok"}
	r := New(exec, 1)
	m := &manifest.Manifest{Calls: []manifest.Call{makeCall("c1", 2)}}
	s := r.Run(context.Background(), m)
	if s.Succeeded != 1 {
		t.Fatalf("expected success after retry, got %+v", s)
	}
	if s.Results[0].Attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", s.Results[0].Attempts)
	}
}

func TestRunner_ExhaustsRetries(t *testing.T) {
	exec := &mockExecutor{failTimes: 99}
	r := New(exec, 1)
	m := &manifest.Manifest{Calls: []manifest.Call{makeCall("c1", 2)}}
	s := r.Run(context.Background(), m)
	if s.Failed != 1 {
		t.Fatalf("expected failure, got %+v", s)
	}
}

func TestRunner_ConcurrencyLimit(t *testing.T) {
	var active atomic.Int32
	var maxSeen atomic.Int32

	exec := &funcExecutor{fn: func(_ context.Context, _ manifest.Call) (string, error) {
		cur := active.Add(1)
		if int(cur) > int(maxSeen.Load()) {
			maxSeen.Store(cur)
		}
		time.Sleep(10 * time.Millisecond)
		active.Add(-1)
		return "ok", nil
	}}

	calls := make([]manifest.Call, 6)
	for i := range calls {
		calls[i] = makeCall("c", 0)
	}
	r := New(exec, 2)
	m := &manifest.Manifest{Calls: calls}
	r.Run(context.Background(), m)

	if maxSeen.Load() > 2 {
		t.Fatalf("concurrency exceeded limit: max active = %d", maxSeen.Load())
	}
}

type funcExecutor struct {
	fn func(context.Context, manifest.Call) (string, error)
}

func (f *funcExecutor) Execute(ctx context.Context, c manifest.Call) (string, error) {
	return f.fn(ctx, c)
}
