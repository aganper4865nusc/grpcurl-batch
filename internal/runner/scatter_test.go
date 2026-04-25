package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func scatterExec(delay time.Duration, err error) func(context.Context, Call) (*Result, error) {
	return func(ctx context.Context, c Call) (*Result, error) {
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		if err != nil {
			return nil, err
		}
		return &Result{Call: c, Err: nil}, nil
	}
}

func TestScatterPolicy_NewEmptyAddresses_Errors(t *testing.T) {
	_, err := NewScatterPolicy(nil)
	if err == nil {
		t.Fatal("expected error for empty address list")
	}
}

func TestScatterPolicy_AllSucceed(t *testing.T) {
	addrs := []string{"host1:50051", "host2:50051", "host3:50051"}
	sp, err := NewScatterPolicy(addrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := Call{Method: "/svc/Method"}
	results := sp.Scatter(context.Background(), call, scatterExec(0, nil))

	if len(results) != len(addrs) {
		t.Fatalf("expected %d results, got %d", len(addrs), len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] unexpected error: %v", i, r.Err)
		}
		if r.Address != addrs[i] {
			t.Errorf("result[%d] address: want %s, got %s", i, addrs[i], r.Address)
		}
	}
}

func TestScatterPolicy_AllFail_FirstSuccessErrors(t *testing.T) {
	sp, _ := NewScatterPolicy([]string{"a:1", "b:1"})
	sentinel := errors.New("boom")
	results := sp.Scatter(context.Background(), Call{}, scatterExec(0, sentinel))

	_, err := FirstSuccess(results)
	if err == nil {
		t.Fatal("expected error when all fail")
	}
}

func TestScatterPolicy_PartialSuccess_FirstSuccessReturns(t *testing.T) {
	addrs := []string{"ok:1", "fail:1"}
	sp, _ := NewScatterPolicy(addrs)

	execFn := func(_ context.Context, c Call) (*Result, error) {
		if c.Address == "fail:1" {
			return nil, errors.New("fail")
		}
		return &Result{Call: c}, nil
	}

	results := sp.Scatter(context.Background(), Call{}, execFn)
	sr, err := FirstSuccess(results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sr.Address != "ok:1" {
		t.Errorf("expected address ok:1, got %s", sr.Address)
	}
}

func TestScatterPolicy_ContextCancelled_SubCallsAbort(t *testing.T) {
	sp, _ := NewScatterPolicy([]string{"a:1", "b:1"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := sp.Scatter(ctx, Call{}, scatterExec(100*time.Millisecond, nil))
	for _, r := range results {
		if r.Err == nil {
			t.Error("expected context cancellation error")
		}
	}
}
