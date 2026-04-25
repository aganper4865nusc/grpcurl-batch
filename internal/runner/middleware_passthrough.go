package runner

import "strings"

// WithPassthrough wraps the provided next CallFunc so that any call whose
// method contains the given suffix (e.g. "/Health" or "Check") is executed
// directly via the executor, bypassing the rest of the middleware chain.
//
// Example:
//
//	chain := Chain(
//		WithRateLimit(rl),
//		WithCircuitBreaker(cb),
//		WithPassthrough(executor, "/Health"),
//	)
func WithPassthrough(direct CallFunc, methodSuffix string) Middleware {
	predicate := func(c Call) bool {
		if methodSuffix == "" {
			return false
		}
		return strings.HasSuffix(c.Method, methodSuffix)
	}
	policy := NewPassthroughPolicy(predicate)
	return func(next CallFunc) CallFunc {
		return policy.Wrap(direct, next)
	}
}

// PassthroughStatus returns a snapshot of bypass statistics for the given
// PassthroughPolicy, suitable for logging or metrics export.
type PassthroughStatus struct {
	Bypassed int64 `json:"bypassed"`
}

// PassthroughSnapshot captures the current bypass count from policy.
func PassthroughSnapshot(p *PassthroughPolicy) PassthroughStatus {
	return PassthroughStatus{Bypassed: p.Bypassed()}
}
