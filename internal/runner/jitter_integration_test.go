package runner

import (
	"sync"
	"testing"
	"time"
)

func TestJitterPolicy_ConcurrentSafety(t *testing.T) {
	base := 50 * time.Millisecond
	j := NewJitterPolicy(fixedBackoff(base), 0.3)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(attempt int) {
			defer wg.Done()
			d := j.Delay(attempt)
			if d < 0 {
				t.Errorf("negative delay in goroutine: %v", d)
			}
		}(i % 5)
	}
	wg.Wait()
}

func TestJitterPolicy_ExponentialBase_StaysBounded(t *testing.T) {
	exp := &backoffPolicy{
		kind:     backoffExponential,
		delay:    10 * time.Millisecond,
		maxDelay: 200 * time.Millisecond,
	}
	j := NewJitterPolicy(exp, 0.2)
	for attempt := 0; attempt <= 10; attempt++ {
		d := j.Delay(attempt)
		max := time.Duration(float64(200*time.Millisecond) * 1.2)
		if d > max {
			t.Errorf("attempt %d: delay %v exceeds expected max %v", attempt, d, max)
		}
		if d < 0 {
			t.Errorf("attempt %d: negative delay %v", attempt, d)
		}
	}
}

// TestJitterPolicy_ZeroFactor ensures that a jitter factor of 0 returns the
// base delay exactly, with no random component applied.
func TestJitterPolicy_ZeroFactor(t *testing.T) {
	base := 100 * time.Millisecond
	j := NewJitterPolicy(fixedBackoff(base), 0.0)
	for attempt := 0; attempt < 5; attempt++ {
		d := j.Delay(attempt)
		if d != base {
			t.Errorf("attempt %d: expected %v with zero jitter, got %v", attempt, base, d)
		}
	}
}
