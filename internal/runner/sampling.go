package runner

import (
	"math/rand"
	"sync"
)

// SamplingPolicy controls what fraction of calls are actually executed.
// Calls that are not sampled receive a nil result and no error, allowing
// the caller to skip them silently.
type SamplingPolicy struct {
	mu   sync.Mutex
	rate float64 // 0.0–1.0
	rng  *rand.Rand
}

// NewSamplingPolicy returns a SamplingPolicy that executes calls with
// probability rate (clamped to [0.0, 1.0]). A rate of 1.0 runs every
// call; a rate of 0.0 skips all calls.
func NewSamplingPolicy(rate float64, src rand.Source) *SamplingPolicy {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	if src == nil {
		src = rand.NewSource(42)
	}
	return &SamplingPolicy{
		rate: rate,
		rng:  rand.New(src),
	}
}

// Sampled reports whether the next call should be executed.
func (s *SamplingPolicy) Sampled() bool {
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0.0 {
		return false
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sampling rate.
func (s *SamplingPolicy) Rate() float64 {
	return s.rate
}
