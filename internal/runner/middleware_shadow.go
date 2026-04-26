package runner

import (
	"context"
	"fmt"
	"time"
)

// WithShadow wraps a call executor with shadow traffic mirroring.
// The primary call is executed normally; a shadow copy is fired asynchronously
// to the shadow address without affecting the primary result or latency.
//
// Usage:
//
//	exec := WithShadow(policy, primaryExec)
//	result, err := exec(ctx, call)
func WithShadow(policy *ShadowPolicy, next func(context.Context, Call) (Result, error)) func(context.Context, Call) (Result, error) {
	if policy == nil {
		return next
	}
	return func(ctx context.Context, call Call) (Result, error) {
		return policy.Execute(ctx, call, next)
	}
}

// ShadowSnapshot holds a point-in-time view of shadow traffic statistics.
type ShadowSnapshot struct {
	// TotalFired is the number of shadow calls dispatched.
	TotalFired int
	// TotalSucceeded is the number of shadow calls that returned no error.
	TotalSucceeded int
	// TotalFailed is the number of shadow calls that returned an error.
	TotalFailed int
	// AvgLatencyMs is the mean shadow call latency in milliseconds.
	AvgLatencyMs float64
	// Results contains the most-recently recorded shadow results (capped by policy.MaxResults).
	Results []ShadowResult
}

// ShadowStatus returns a snapshot of the shadow policy's recorded results.
// It is safe to call concurrently.
func ShadowStatus(policy *ShadowPolicy) ShadowSnapshot {
	if policy == nil {
		return ShadowSnapshot{}
	}

	results := policy.Results()
	var succeeded, failed int
	var totalMs float64

	for _, r := range results {
		if r.Err == nil {
			succeeded++
		} else {
			failed++
		}
		totalMs += float64(r.Latency.Milliseconds())
	}

	var avgMs float64
	if len(results) > 0 {
		avgMs = totalMs / float64(len(results))
	}

	return ShadowSnapshot{
		TotalFired:     len(results),
		TotalSucceeded: succeeded,
		TotalFailed:    failed,
		AvgLatencyMs:   avgMs,
		Results:        results,
	}
}

// FormatShadowSnapshot renders a ShadowSnapshot as a human-readable string
// suitable for logging or CLI output.
func FormatShadowSnapshot(s ShadowSnapshot) string {
	if s.TotalFired == 0 {
		return "shadow: no calls recorded"
	}
	return fmt.Sprintf(
		"shadow: fired=%d succeeded=%d failed=%d avg_latency=%.1fms",
		s.TotalFired, s.TotalSucceeded, s.TotalFailed, s.AvgLatencyMs,
	)
}

// ShadowResult is a single recorded outcome from a shadow call.
// Re-exported here so callers do not need to import shadow.go internals directly.
type shadowResultAlias = ShadowResult

// shadowResultAlias is intentionally unexported; callers use ShadowResult from shadow.go.
// This file provides middleware helpers only.

// shadowDuration is a helper used in tests to measure shadow call durations
// against a fixed reference point.
func shadowDuration(start time.Time) time.Duration {
	return time.Since(start)
}
