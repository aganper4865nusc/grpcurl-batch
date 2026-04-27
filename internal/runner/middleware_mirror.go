package runner

import (
	"context"
	"fmt"
)

// WithMirror wraps next with a MirrorPolicy that fans out calls to the
// supplied mirror addresses. Mirror results are discarded; only the
// primary outcome is returned to the caller.
//
// The executor used for mirror calls is the same next function with the
// call address rewritten, so it participates in the full middleware chain
// downstream of this point.
func WithMirror(p *MirrorPolicy) Middleware {
	return func(next CallFunc) CallFunc {
		// Supply p.exec with a closure that delegates back to next so that
		// mirror calls benefit from retry/timeout layers below this one.
		p.exec = func(ctx context.Context, call Call) (string, error) {
			return next(ctx, call)
		}
		return p.Wrap(next)
	}
}

// MirrorStatus summarises the current state of a MirrorPolicy for
// display in status output or health endpoints.
type MirrorStatus struct {
	Mirrors []string `json:"mirrors"`
	Count   int      `json:"count"`
}

// MirrorSnapshot returns a MirrorStatus snapshot of p.
func MirrorSnapshot(p *MirrorPolicy) MirrorStatus {
	m := p.Mirrors()
	return MirrorStatus{Mirrors: m, Count: len(m)}
}

// FormatMirrorStatus returns a human-readable summary of the mirror pool.
func FormatMirrorStatus(s MirrorStatus) string {
	if s.Count == 0 {
		return "mirror: no mirrors configured"
	}
	return fmt.Sprintf("mirror: %d active mirror(s): %v", s.Count, s.Mirrors)
}
