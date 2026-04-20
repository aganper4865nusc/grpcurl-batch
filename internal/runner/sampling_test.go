package runner

import (
	"context"
	"math/rand"
	"testing"

	"github.com/yourorg/grpcurl-batch/internal/manifest"
)

func TestSamplingPolicy_FullRate_AlwaysSampled(t *testing.T) {
	p := NewSamplingPolicy(1.0, rand.NewSource(0))
	for i := 0; i < 100; i++ {
		if !p.Sampled() {
			t.Fatal("expected sampled at rate 1.0")
		}
	}
}

func TestSamplingPolicy_ZeroRate_NeverSampled(t *testing.T) {
	p := NewSamplingPolicy(0.0, rand.NewSource(0))
	for i := 0; i < 100; i++ {
		if p.Sampled() {
			t.Fatal("expected not sampled at rate 0.0")
		}
	}
}

func TestSamplingPolicy_ClampsBelowZero(t *testing.T) {
	p := NewSamplingPolicy(-5.0, nil)
	if p.Rate() != 0.0 {
		t.Fatalf("expected rate 0.0, got %v", p.Rate())
	}
}

func TestSamplingPolicy_ClampsAboveOne(t *testing.T) {
	p := NewSamplingPolicy(3.0, nil)
	if p.Rate() != 1.0 {
		t.Fatalf("expected rate 1.0, got %v", p.Rate())
	}
}

func TestSamplingPolicy_HalfRate_RoughlyHalf(t *testing.T) {
	p := NewSamplingPolicy(0.5, rand.NewSource(99))
	count := 0
	const n = 10_000
	for i := 0; i < n; i++ {
		if p.Sampled() {
			count++
		}
	}
	got := float64(count) / n
	if got < 0.45 || got > 0.55 {
		t.Fatalf("expected ~0.5 sample rate, got %.3f", got)
	}
}

func TestWithSampling_SkipsCall(t *testing.T) {
	p := NewSamplingPolicy(0.0, nil)
	mw := WithSampling(p)
	called := false
	next := func(_ context.Context, _ manifest.Call) (string, error) {
		called = true
		return "ok", nil
	}
	_, err := mw(next)(context.Background(), manifest.Call{})
	if err != ErrCallSkipped {
		t.Fatalf("expected ErrCallSkipped, got %v", err)
	}
	if called {
		t.Fatal("next should not have been called")
	}
}

func TestWithSampling_AllowsCall(t *testing.T) {
	p := NewSamplingPolicy(1.0, nil)
	mw := WithSampling(p)
	next := func(_ context.Context, _ manifest.Call) (string, error) {
		return "result", nil
	}
	out, err := mw(next)(context.Background(), manifest.Call{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "result" {
		t.Fatalf("expected 'result', got %q", out)
	}
}
