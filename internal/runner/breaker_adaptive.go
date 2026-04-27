package runner

import (
	"errors"
	"sync"
	"time"
)

// ErrAdaptiveBreakerOpen is returned when the adaptive circuit breaker is open.
var ErrAdaptiveBreakerOpen = errors.New("adaptive circuit breaker: circuit open")

// AdaptiveBreakerConfig holds tuning parameters for the adaptive circuit breaker.
type AdaptiveBreakerConfig struct {
	// MinRequests is the minimum number of calls before the breaker can open.
	MinRequests int
	// ErrorRateThreshold is the fraction [0,1] of failures that triggers opening.
	ErrorRateThreshold float64
	// HalfOpenAfter is how long to wait before trying a probe call.
	HalfOpenAfter time.Duration
	// WindowSize is the rolling window used to measure error rate.
	WindowSize time.Duration
}

type adaptiveBreakerState int

const (
	abClosed  adaptiveBreakerState = iota
	abOpen
	abHalfOpen
)

// AdaptiveBreaker is a circuit breaker whose open/close decision is driven by
// a sliding-window error rate rather than a fixed failure count.
type AdaptiveBreaker struct {
	cfg      AdaptiveBreakerConfig
	mu       sync.Mutex
	state    adaptiveBreakerState
	openedAt time.Time
	window   *WindowPolicy
}

// NewAdaptiveBreaker creates an AdaptiveBreaker with the given config.
func NewAdaptiveBreaker(cfg AdaptiveBreakerConfig) *AdaptiveBreaker {
	if cfg.MinRequests <= 0 {
		cfg.MinRequests = 5
	}
	if cfg.ErrorRateThreshold <= 0 {
		cfg.ErrorRateThreshold = 0.5
	}
	if cfg.HalfOpenAfter <= 0 {
		cfg.HalfOpenAfter = 10 * time.Second
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 30 * time.Second
	}
	return &AdaptiveBreaker{
		cfg:    cfg,
		state:  abClosed,
		window: NewWindowPolicy(cfg.WindowSize),
	}
}

// Allow returns nil if the call should proceed, or ErrAdaptiveBreakerOpen.
func (ab *AdaptiveBreaker) Allow() error {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	switch ab.state {
	case abOpen:
		if time.Since(ab.openedAt) >= ab.cfg.HalfOpenAfter {
			ab.state = abHalfOpen
			return nil
		}
		return ErrAdaptiveBreakerOpen
	default:
		return nil
	}
}

// Record registers the outcome of a call and potentially trips or resets the breaker.
func (ab *AdaptiveBreaker) Record(success bool) {
	ab.window.Record(success)
	ab.mu.Lock()
	defer ab.mu.Unlock()
	stats := ab.window.Stats()
	switch ab.state {
	case abHalfOpen:
		if success {
			ab.state = abClosed
		} else {
			ab.state = abOpen
			ab.openedAt = time.Now()
		}
	case abClosed:
		if stats.Total >= ab.cfg.MinRequests && stats.ErrorRate() >= ab.cfg.ErrorRateThreshold {
			ab.state = abOpen
			ab.openedAt = time.Now()
		}
	}
}

// State returns the current breaker state as a string.
func (ab *AdaptiveBreaker) State() string {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	switch ab.state {
	case abOpen:
		return "open"
	case abHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}
