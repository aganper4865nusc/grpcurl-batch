package runner

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ProbeResult holds the outcome of a single readiness probe.
type ProbeResult struct {
	Healthy  bool
	Latency  time.Duration
	Err      error
	CheckedAt time.Time
}

// ProbeFunc is a function that checks whether a target is reachable.
type ProbeFunc func(ctx context.Context, address string) error

// ProbePolicy gates call execution behind a readiness check.
type ProbePolicy struct {
	probe      ProbeFunc
	timeout    time.Duration
	maxFails   int32
	failCount  atomic.Int32
	lastResult atomic.Pointer[ProbeResult]
}

// NewProbePolicy creates a ProbePolicy with the given probe function.
// timeout is applied to each probe invocation; maxFails consecutive failures
// cause subsequent Wrap calls to return an error immediately.
func NewProbePolicy(probe ProbeFunc, timeout time.Duration, maxFails int) *ProbePolicy {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if maxFails <= 0 {
		maxFails = 3
	}
	return &ProbePolicy{
		probe:    probe,
		timeout:  timeout,
		maxFails: int32(maxFails),
	}
}

// Check runs the probe against address and records the result.
func (p *ProbePolicy) Check(ctx context.Context, address string) ProbeResult {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	start := time.Now()
	err := p.probe(ctx, address)
	res := ProbeResult{
		Healthy:   err == nil,
		Latency:   time.Since(start),
		Err:       err,
		CheckedAt: time.Now(),
	}
	p.lastResult.Store(&res)
	if err != nil {
		p.failCount.Add(1)
	} else {
		p.failCount.Store(0)
	}
	return res
}

// Wrap executes fn only if the target passes the readiness probe.
func (p *ProbePolicy) Wrap(ctx context.Context, address string, fn func(ctx context.Context) (string, error)) (string, error) {
	if p.failCount.Load() >= p.maxFails {
		return "", fmt.Errorf("probe: target %q is unhealthy after %d consecutive failures", address, p.failCount.Load())
	}
	res := p.Check(ctx, address)
	if !res.Healthy {
		return "", fmt.Errorf("probe: target %q not ready: %w", address, res.Err)
	}
	return fn(ctx)
}

// LastResult returns the most recent probe result, or nil if no probe has run.
func (p *ProbePolicy) LastResult() *ProbeResult {
	return p.lastResult.Load()
}
