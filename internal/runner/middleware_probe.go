package runner

import (
	"context"
	"fmt"
	"time"
)

// ProbeStatus holds a snapshot of the probe policy state.
type ProbeStatus struct {
	Target    string
	Healthy   bool
	FailCount int
	LastCheck time.Time
}

// WithProbe wraps a call with a health-probe gate. If the probe reports
// the target as unhealthy the call is rejected immediately; otherwise it
// is forwarded to next. The probe is evaluated before every call.
func WithProbe(p *ProbePolicy) func(CallFunc) CallFunc {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, call Call) (string, error) {
			if err := p.Check(ctx, call); err != nil {
				return "", fmt.Errorf("probe gate: %w", err)
			}
			return next(ctx, call)
		}
	}
}

// ProbeSnapshot returns a point-in-time view of the probe policy.
func ProbeSnapshot(p *ProbePolicy) ProbeStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ProbeStatus{
		Target:    p.target,
		Healthy:   p.healthy,
		FailCount: p.failCount,
		LastCheck: p.lastCheck,
	}
}
