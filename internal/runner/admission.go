package runner

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrAdmissionDenied is returned when a call is rejected by the admission policy.
var ErrAdmissionDenied = errors.New("admission denied: server overloaded")

// AdmissionPolicy gates calls based on a dynamic load score between 0.0 and 1.0.
// When the load score exceeds the threshold, new calls are rejected.
type AdmissionPolicy struct {
	threshold float64
	loadScore atomic.Value // stores float64
}

// NewAdmissionPolicy creates an AdmissionPolicy with the given rejection threshold (0.0–1.0).
// A threshold of 1.0 never rejects; 0.0 always rejects.
func NewAdmissionPolicy(threshold float64) *AdmissionPolicy {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	ap := &AdmissionPolicy{threshold: threshold}
	ap.loadScore.Store(float64(0))
	return ap
}

// SetLoad updates the current load score (0.0 = idle, 1.0 = fully loaded).
func (ap *AdmissionPolicy) SetLoad(score float64) {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	ap.loadScore.Store(score)
}

// Load returns the current load score.
func (ap *AdmissionPolicy) Load() float64 {
	return ap.loadScore.Load().(float64)
}

// Admit returns nil if the call should proceed, or ErrAdmissionDenied if rejected.
func (ap *AdmissionPolicy) Admit() error {
	if ap.Load() >= ap.threshold {
		return ErrAdmissionDenied
	}
	return nil
}

// Wrap executes fn only if the current load is below the threshold.
func (ap *AdmissionPolicy) Wrap(ctx context.Context, fn func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ap.Admit(); err != nil {
		return err
	}
	return fn(ctx)
}
