package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func mirrorCall(addr string) Call {
	return Call{Address: addr, Service: "svc", Method: "Ping"}
}

func TestMirrorPolicy_PrimarySucceeds_ReturnsPrimaryResult(t *testing.T) {
	exec := func(_ context.Context, _ Call) (string, error) { return "", nil }
	p := NewMirrorPolicy([]string{"mirror:9000"}, exec)
	next := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	res, err := p.Wrap(next)(context.Background(), mirrorCall("primary:9000"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "ok" {
		t.Fatalf("expected 'ok', got %q", res)
	}
}

func TestMirrorPolicy_PrimaryFails_ReturnsError(t *testing.T) {
	exec := func(_ context.Context, _ Call) (string, error) { return "", nil }
	p := NewMirrorPolicy(nil, exec)
	next := func(_ context.Context, _ Call) (string, error) { return "", errors.New("boom") }
	_, err := p.Wrap(next)(context.Background(), mirrorCall("primary:9000"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMirrorPolicy_MirrorCallsFired(t *testing.T) {
	var fired atomic.Int32
	exec := func(_ context.Context, _ Call) (string, error) {
		fired.Add(1)
		return "", nil
	}
	p := NewMirrorPolicy([]string{"m1:9000", "m2:9000"}, exec)
	next := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	_, _ = p.Wrap(next)(context.Background(), mirrorCall("primary:9000"))
	time.Sleep(20 * time.Millisecond)
	if got := int(fired.Load()); got != 2 {
		t.Fatalf("expected 2 mirror calls, got %d", got)
	}
}

func TestMirrorPolicy_MirrorAddressOverridden(t *testing.T) {
	var gotAddr string
	exec := func(_ context.Context, c Call) (string, error) {
		gotAddr = c.Address
		return "", nil
	}
	p := NewMirrorPolicy([]string{"mirror:9000"}, exec)
	next := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	_, _ = p.Wrap(next)(context.Background(), mirrorCall("primary:9000"))
	time.Sleep(20 * time.Millisecond)
	if gotAddr != "mirror:9000" {
		t.Fatalf("expected mirror:9000, got %q", gotAddr)
	}
}

func TestMirrorPolicy_AddAndRemoveMirror(t *testing.T) {
	p := NewMirrorPolicy([]string{"a:1"}, nil)
	p.AddMirror("b:2")
	if len(p.Mirrors()) != 2 {
		t.Fatalf("expected 2 mirrors")
	}
	p.RemoveMirror("a:1")
	m := p.Mirrors()
	if len(m) != 1 || m[0] != "b:2" {
		t.Fatalf("unexpected mirrors: %v", m)
	}
}

func TestMirrorPolicy_EmptyMirrors_OnlyPrimaryRuns(t *testing.T) {
	var execCalled bool
	exec := func(_ context.Context, _ Call) (string, error) { execCalled = true; return "", nil }
	p := NewMirrorPolicy(nil, exec)
	next := func(_ context.Context, _ Call) (string, error) { return "primary", nil }
	res, err := p.Wrap(next)(context.Background(), mirrorCall("primary:9000"))
	if err != nil || res != "primary" {
		t.Fatalf("unexpected result: %v %v", res, err)
	}
	time.Sleep(10 * time.Millisecond)
	if execCalled {
		t.Fatal("exec should not have been called with no mirrors")
	}
}
