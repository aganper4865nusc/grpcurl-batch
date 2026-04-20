package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isobit/grpcurl-batch/internal/manifest"
)

func okCall(_ context.Context, _ manifest.Call) (string, error) { return "ok", nil }
func errCall(_ context.Context, _ manifest.Call) (string, error) {
	return "", errors.New("fail")
}

func TestWithWindowBreaker_AllowsBelowMinCalls(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	mw := WithWindowBreaker(w, 5, 0.5)
	wrapped := mw(errCall)
	ctx := context.Background()
	c := manifest.Call{}
	for i := 0; i < 4; i++ {
		_, _ = wrapped(ctx, c)
	}
	_, err := wrapped(ctx, c)
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected underlying error before minCalls, got %v", err)
	}
}

func TestWithWindowBreaker_OpensAboveFailRate(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	mw := WithWindowBreaker(w, 3, 0.5)
	wrapped := mw(errCall)
	ctx := context.Background()
	c := manifest.Call{}
	for i := 0; i < 4; i++ {
		_, _ = wrapped(ctx, c)
	}
	_, err := wrapped(ctx, c)
	if err == nil || err.Error() == "fail" {
		t.Fatalf("expected breaker open error, got %v", err)
	}
}

func TestWithWindowBreaker_ClosedWhenSuccessful(t *testing.T) {
	w := NewWindowPolicy(time.Second)
	mw := WithWindowBreaker(w, 3, 0.5)
	wrapped := mw(okCall)
	ctx := context.Background()
	c := manifest.Call{}
	for i := 0; i < 5; i++ {
		_, err := wrapped(ctx, c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}
}
