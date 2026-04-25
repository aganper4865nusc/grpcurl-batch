package runner

import (
	"context"
	"errors"
	"testing"
)

func passThroughCall(method string, tags []string) Call {
	return Call{Method: method, Tags: tags}
}

func TestPassthrough_PredicateFalse_UsesNext(t *testing.T) {
	p := NewPassthroughPolicy(func(c Call) bool { return false })
	nextCalled := false
	next := func(_ context.Context, _ Call) (string, error) {
		nextCalled = true
		return "next", nil
	}
	direct := func(_ context.Context, _ Call) (string, error) { return "direct", nil }

	wrapped := p.Wrap(direct, next)
	out, err := wrapped(context.Background(), passThroughCall("svc/Method", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "next" || !nextCalled {
		t.Errorf("expected next to be called, got %q", out)
	}
	if p.Bypassed() != 0 {
		t.Errorf("expected 0 bypassed, got %d", p.Bypassed())
	}
}

func TestPassthrough_PredicateTrue_UsesDirect(t *testing.T) {
	p := NewPassthroughPolicy(func(c Call) bool { return true })
	direct := func(_ context.Context, _ Call) (string, error) { return "direct", nil }
	next := func(_ context.Context, _ Call) (string, error) { return "next", nil }

	wrapped := p.Wrap(direct, next)
	out, err := wrapped(context.Background(), passThroughCall("svc/Health", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "direct" {
		t.Errorf("expected direct, got %q", out)
	}
	if p.Bypassed() != 1 {
		t.Errorf("expected 1 bypassed, got %d", p.Bypassed())
	}
}

func TestPassthrough_NilPredicate_NeverBypasses(t *testing.T) {
	p := NewPassthroughPolicy(nil)
	nextCalled := false
	next := func(_ context.Context, _ Call) (string, error) {
		nextCalled = true
		return "next", nil
	}
	direct := func(_ context.Context, _ Call) (string, error) { return "direct", nil }

	wrapped := p.Wrap(direct, next)
	_, _ = wrapped(context.Background(), passThroughCall("svc/Method", nil))
	if !nextCalled {
		t.Error("expected next to be called with nil predicate")
	}
}

func TestPassthrough_CancelledContext_ReturnsErr(t *testing.T) {
	p := NewPassthroughPolicy(func(c Call) bool { return true })
	direct := func(_ context.Context, _ Call) (string, error) { return "ok", nil }
	next := func(_ context.Context, _ Call) (string, error) { return "ok", nil }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	wrapped := p.Wrap(direct, next)
	_, err := wrapped(ctx, passThroughCall("svc/Method", nil))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPassthrough_NilDirect_FallsBackToNext(t *testing.T) {
	p := NewPassthroughPolicy(func(c Call) bool { return true })
	next := func(_ context.Context, _ Call) (string, error) { return "next", nil }

	wrapped := p.Wrap(nil, next)
	out, err := wrapped(context.Background(), passThroughCall("svc/Method", nil))
	if err != nil || out != "next" {
		t.Errorf("expected next fallback, got %q %v", out, err)
	}
}

func TestPassthrough_BypassedCounterAccumulates(t *testing.T) {
	p := NewPassthroughPolicy(func(c Call) bool { return true })
	direct := func(_ context.Context, _ Call) (string, error) { return "", nil }
	next := func(_ context.Context, _ Call) (string, error) { return "", nil }
	wrapped := p.Wrap(direct, next)

	for i := 0; i < 5; i++ {
		_, _ = wrapped(context.Background(), passThroughCall("svc/Method", nil))
	}
	if p.Bypassed() != 5 {
		t.Errorf("expected 5 bypassed, got %d", p.Bypassed())
	}
}
