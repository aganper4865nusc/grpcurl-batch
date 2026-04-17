package runner

import (
	"testing"
	"time"
)

func fixedBackoff(d time.Duration) BackoffPolicy {
	return &backoffPolicy{
		kind:     backoffFixed,
		delay:    d,
		maxDelay: d,
	}
}

func TestJitterPolicy_ZeroBase_ReturnsZero(t *testing.T) {
	j := NewJitterPolicy(fixedBackoff(0), 0.2)
	if got := j.Delay(1); got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestJitterPolicy_ZeroFactor_ReturnsBase(t *testing.T) {
	base := 100 * time.Millisecond
	j := NewJitterPolicy(fixedBackoff(base), 0)
	if got := j.Delay(1); got != base {
		t.Errorf("expected %v, got %v", base, got)
	}
}

func TestJitterPolicy_DelayWithinBounds(t *testing.T) {
	base := 200 * time.Millisecond
	factor := 0.25
	j := NewJitterPolicy(fixedBackoff(base), factor)
	window := time.Duration(float64(base) * factor)
	lo := base - window
	hi := base + window
	for i := 0; i < 100; i++ {
		d := j.Delay(1)
		if d < lo || d > hi {
			t.Errorf("delay %v out of [%v, %v]", d, lo, hi)
		}
	}
}

func TestJitterPolicy_FactorClamped(t *testing.T) {
	base := 100 * time.Millisecond
	j := NewJitterPolicy(fixedBackoff(base), 5.0) // should clamp to 1.0
	for i := 0; i < 50; i++ {
		d := j.Delay(1)
		if d < 0 {
			t.Errorf("negative delay: %v", d)
		}
		if d > 2*base {
			t.Errorf("delay %v exceeds 2x base %v", d, base)
		}
	}
}

func TestJitterPolicy_NeverNegative(t *testing.T) {
	base := 1 * time.Millisecond
	j := NewJitterPolicy(fixedBackoff(base), 1.0)
	for i := 0; i < 200; i++ {
		if d := j.Delay(1); d < 0 {
			t.Errorf("got negative delay: %v", d)
		}
	}
}
