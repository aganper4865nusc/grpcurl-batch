package runner

// WithCoalesce wraps next with a CoalescePolicy so that concurrent
// identical calls (same deduplication key) share a single in-flight
// execution. keyFn may be nil to use DefaultDedupeKey.
//
// Example usage:
//
//	chain := Chain(base, WithCoalesce(nil))
func WithCoalesce(keyFn func(Call) string) Middleware {
	policy := NewCoalescePolicy(keyFn)
	return func(next CallFunc) CallFunc {
		return policy.Wrap(next)
	}
}

// CoalesceStatus returns a snapshot of the current in-flight count
// tracked by the given CoalescePolicy. Useful for metrics/health checks.
func CoalesceStatus(cp *CoalescePolicy) map[string]int {
	return map[string]int{
		"inflight": cp.Inflight(),
	}
}
