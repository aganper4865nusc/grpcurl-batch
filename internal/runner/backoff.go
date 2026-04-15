package runner

import (
	"math"
	"time"
)

// BackoffStrategy defines how delay is calculated between retry attempts.
type BackoffStrategy int

const (
	// BackoffFixed uses a constant delay between attempts.
	BackoffFixed BackoffStrategy = iota
	// BackoffExponential doubles the delay on each attempt.
	BackoffExponential
)

// BackoffPolicy calculates the wait duration before a retry attempt.
type BackoffPolicy struct {
	Strategy   BackoffStrategy
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

// DefaultBackoffPolicy returns a sensible exponential backoff configuration.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  200 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 2.0,
	}
}

// Delay returns the wait duration for the given attempt number (0-indexed).
func (b BackoffPolicy) Delay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	switch b.Strategy {
	case BackoffExponential:
		multiplier := b.Multiplier
		if multiplier <= 0 {
			multiplier = 2.0
		}
		delay := float64(b.BaseDelay) * math.Pow(multiplier, float64(attempt-1))
		d := time.Duration(delay)
		if b.MaxDelay > 0 && d > b.MaxDelay {
			return b.MaxDelay
		}
		return d
	default: // BackoffFixed
		return b.BaseDelay
	}
}
