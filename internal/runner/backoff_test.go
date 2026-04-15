package runner

import (
	"testing"
	"time"
)

func TestBackoffPolicy_Fixed_ConstantDelay(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffFixed,
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  5 * time.Second,
	}
	for attempt := 1; attempt <= 5; attempt++ {
		got := p.Delay(attempt)
		if got != 100*time.Millisecond {
			t.Errorf("attempt %d: expected 100ms, got %v", attempt, got)
		}
	}
}

func TestBackoffPolicy_Exponential_Doubles(t *testing.T) {
	p := BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   0,
		Multiplier: 2.0,
	}
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		got := p.Delay(i + 1)
		if got != want {
			t.Errorf("attempt %d: expected %v, got %v", i+1, want, got)
		}
	}
}

func TestBackoffPolicy_Exponential_RespectsMaxDelay(t *testing.T) {
	p := BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
	}
	for attempt := 3; attempt <= 6; attempt++ {
		got := p.Delay(attempt)
		if got > 1*time.Second {
			t.Errorf("attempt %d: delay %v exceeds MaxDelay 1s", attempt, got)
		}
	}
}

func TestBackoffPolicy_ZeroAttempt_ReturnsZero(t *testing.T) {
	p := DefaultBackoffPolicy()
	if d := p.Delay(0); d != 0 {
		t.Errorf("expected 0 delay for attempt 0, got %v", d)
	}
}

func TestDefaultBackoffPolicy_IsExponential(t *testing.T) {
	p := DefaultBackoffPolicy()
	if p.Strategy != BackoffExponential {
		t.Errorf("expected BackoffExponential strategy")
	}
	if p.BaseDelay != 200*time.Millisecond {
		t.Errorf("expected 200ms base delay, got %v", p.BaseDelay)
	}
	if p.MaxDelay != 10*time.Second {
		t.Errorf("expected 10s max delay, got %v", p.MaxDelay)
	}
}

func TestBackoffPolicy_DefaultMultiplier_WhenZero(t *testing.T) {
	p := BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  100 * time.Millisecond,
		Multiplier: 0,
	}
	// With zero multiplier, should default to 2.0 internally
	d1 := p.Delay(1)
	d2 := p.Delay(2)
	if d2 != d1*2 {
		t.Errorf("expected doubling with zero multiplier fallback, got d1=%v d2=%v", d1, d2)
	}
}
