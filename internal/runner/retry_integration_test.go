package runner_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/runner"
)

// TestRetryPolicy_DelayBetweenAttempts verifies that the configured delay is
// actually observed between consecutive attempts.
func TestRetryPolicy_DelayBetweenAttempts(t *testing.T) {
	const delay = 50 * time.Millisecond
	p := runner.RetryPolicy{MaxAttempts: 3, Delay: delay}

	var timestamps []time.Time
	_ = p.Execute(context.Background(), func(_ context.Context, _ int) error {
		timestamps = append(timestamps, time.Now())
		return errors.New("always fail")
	})

	if len(timestamps) != 3 {
		t.Fatalf("expected 3 timestamps, got %d", len(timestamps))
	}

	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		if gap < delay {
			t.Errorf("gap between attempt %d and %d was %v, want >= %v", i, i+1, gap, delay)
		}
	}
}

// TestRetryPolicy_ConcurrentSafety ensures RetryPolicy can be shared across
// goroutines without data races.
func TestRetryPolicy_ConcurrentSafety(t *testing.T) {
	p := runner.RetryPolicy{MaxAttempts: 5, Delay: 0}

	var successCount atomic.Int64
	const workers = 20

	done := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			err := p.Execute(context.Background(), func(_ context.Context, _ int) error {
				return nil
			})
			if err == nil {
				successCount.Add(1)
			}
		}()
	}

	for i := 0; i < workers; i++ {
		<-done
	}

	if got := successCount.Load(); got != workers {
		t.Errorf("expected %d successes, got %d", workers, got)
	}
}
