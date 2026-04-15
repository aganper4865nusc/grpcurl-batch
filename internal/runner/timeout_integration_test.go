package runner_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/runner"
)

func TestTimeoutPolicy_ConcurrentCalls(t *testing.T) {
	tp := runner.TimeoutPolicy{Timeout: 200 * time.Millisecond}
	const goroutines = 20

	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = tp.Apply(context.Background(), func(_ context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error %v", i, err)
		}
	}
}

func TestTimeoutPolicy_SlowCallsTimedOut(t *testing.T) {
	tp := runner.TimeoutPolicy{Timeout: 50 * time.Millisecond}
	const goroutines = 10

	var wg sync.WaitGroup
	timeouts := 0
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := tp.Apply(context.Background(), func(ctx context.Context) error {
				select {
				case <-time.After(200 * time.Millisecond):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			if errors.Is(err, context.DeadlineExceeded) {
				mu.Lock()
				timeouts++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if timeouts != goroutines {
		t.Fatalf("expected %d timeouts, got %d", goroutines, timeouts)
	}
}
