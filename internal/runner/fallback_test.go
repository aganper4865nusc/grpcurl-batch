package runner

import (
	"context"
	"errors"
	"testing"
)

var errPrimary = errors.New("primary failed")

func primaryOK(_ context.Context) (string, error)  { return "ok", nil }
func primaryFail(_ context.Context) (string, error) { return "", errPrimary }

func TestFallbackPolicy_PrimarySucceeds_NoFallback(t *testing.T) {
	invoked := false
	fp := NewFallbackPolicy(func(_ context.Context, _ Call, _ error) (string, error) {
		invoked = true
		return "fallback", nil
	})
	res, err := fp.Execute(context.Background(), Call{}, primaryOK)
	if err != nil || res != "ok" || invoked {
		t.Fatalf("expected ok without fallback, got res=%q err=%v invoked=%v", res, err, invoked)
	}
}

func TestFallbackPolicy_PrimaryFails_FallbackInvoked(t *testing.T) {
	fp := NewFallbackPolicy(func(_ context.Context, _ Call, err error) (string, error) {
		if !errors.Is(err, errPrimary) {
			return "", errors.New("unexpected error")
		}
		return "fallback-response", nil
	})
	res, err := fp.Execute(context.Background(), Call{}, primaryFail)
	if err != nil || res != "fallback-response" {
		t.Fatalf("expected fallback-response, got res=%q err=%v", res, err)
	}
}

func TestFallbackPolicy_ContextCancelled_SkipsFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	invoked := false
	fp := NewFallbackPolicy(func(_ context.Context, _ Call, _ error) (string, error) {
		invoked = true
		return "fb", nil
	})
	_, err := fp.Execute(ctx, Call{}, func(c context.Context) (string, error) {
		return "", c.Err()
	})
	if !errors.Is(err, context.Canceled) || invoked {
		t.Fatalf("expected context.Canceled without fallback, got err=%v invoked=%v", err, invoked)
	}
}

func TestFallbackPolicy_NilFallbackFn_ReturnsOriginalError(t *testing.T) {
	fp := NewFallbackPolicy(nil)
	_, err := fp.Execute(context.Background(), Call{}, primaryFail)
	if !errors.Is(err, errPrimary) {
		t.Fatalf("expected errPrimary, got %v", err)
	}
}

func TestStaticFallback_ReturnsFixedPayload(t *testing.T) {
	fp := StaticFallback("default")
	res, err := fp.Execute(context.Background(), Call{}, primaryFail)
	if err != nil || res != "default" {
		t.Fatalf("expected default payload, got res=%q err=%v", res, err)
	}
}
