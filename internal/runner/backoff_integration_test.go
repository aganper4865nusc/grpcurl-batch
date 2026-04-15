package runner

import (
	"testing"
	"time"
)

// TestBackoffPolicy_DelaysGrowOverAttempts verifies that exponential delays
// are strictly increasing across consecutive attempts up to MaxDelay.
func TestBackoffPolicy_DelaysGrowOverAttempts(t *testing.T) {
	p := BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   2 * time.Second,
		Multiplier: 2.0,
	}
	prev := p.Delay(1)
	for attempt := 2; attempt <= 6; attempt++ {
		curr := p.Delay(attempt)
		if curr <= prev && curr < p.MaxDelay {
			t.Errorf("attempt %d: delay %v should be greater than prev %v", attempt, curr, prev)
		}
		prev = curr
	}
}

// TestBackoffPolicy_FixedDoesNotGrow verifies fixed strategy stays constant.
func TestBackoffPolicy_FixedDoesNotGrow(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffFixed,
		BaseDelay: 150 * time.Millisecond,
		MaxDelay:  5 * time.Second,
	}
	base := p.Delay(1)
	for attempt := 2; attempt <= 8; attempt++ {
		if d := p.Delay(attempt); d != base {
			t.Errorf("attempt %d: expected fixed %v, got %v", attempt, base, d)
		}
	}
}

// TestBackoffPolicy_MaxDelay_Clamps ensures no delay exceeds the configured max.
func TestBackoffPolicy_MaxDelay_Clamps(t *testing.T) {
	max := 300 * time.Millisecond
	p := BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   max,
		Multiplier: 3.0,
	}
	for attempt := 1; attempt <= 10; attempt++ {
		if d := p.Delay(attempt); d > max {
			t.Errorf("attempt %d: delay %v exceeds max %v", attempt, d, max)
		}
	}
}
