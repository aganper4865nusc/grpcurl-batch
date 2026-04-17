package runner

import (
	"math/rand"
	"time"
)

// JitterPolicy adds randomized jitter to a base delay to avoid thundering herd.
type JitterPolicy struct {
	base   BackoffPolicy
	factor float64 // fraction of delay to jitter, e.g. 0.2 = ±20%
	rng    *rand.Rand
}

// NewJitterPolicy wraps a BackoffPolicy with jitter.
// factor should be in range (0, 1]. Values outside are clamped.
func NewJitterPolicy(base BackoffPolicy, factor float64) *JitterPolicy {
	if factor <= 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return &JitterPolicy{
		base:   base,
		factor: factor,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Delay returns the base delay with random jitter applied.
func (j *JitterPolicy) Delay(attempt int) time.Duration {
	base := j.base.Delay(attempt)
	if base == 0 || j.factor == 0 {
		return base
	}
	window := float64(base) * j.factor
	// jitter in range [-window, +window]
	offset := (j.rng.Float64()*2 - 1) * window
	result := time.Duration(float64(base) + offset)
	if result < 0 {
		return 0
	}
	return result
}
