package runner

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker prevents repeated calls to a failing endpoint.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        circuitState
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
}

// NewCircuitBreaker creates a CircuitBreaker with the given failure threshold
// and reset timeout. After threshold consecutive failures the circuit opens;
// after resetTimeout it moves to half-open to allow a probe call.
func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 1
	}
	if resetTimeout <= 0 {
		resetTimeout = 10 * time.Second
	}
	return &CircuitBreaker{
		state:        stateClosed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// Allow returns nil if the call is permitted, or ErrCircuitOpen if the
// breaker is open and the reset timeout has not yet elapsed.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = stateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

// RecordFailure increments the failure counter and opens the circuit when
// the threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = stateOpen
		cb.openedAt = time.Now()
	}
}

// State returns the current circuit state as a string (closed/open/half-open).
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}
