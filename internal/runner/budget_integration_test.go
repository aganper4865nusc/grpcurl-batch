package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBudget_ConcurrentRecord_NoDataRace(t *testing.T) {
	p := NewBudgetPolicy(1000, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Record()
			_ = p.Remaining()
		}()
	}
	wg.Wait()
}

func TestBudget_ExhaustedUnderConcurrency(t *testing.T) {
	p := NewBudgetPolicy(5, time.Minute)
	for i := 0; i < 5; i++ {
		p.Record()
	}
	var rejected int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.Wrap(context.Background(), func(_ context.Context) error { return nil })
			if errors.Is(err, ErrBudgetExhausted) {
				mu.Lock()
				rejected++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if rejected != 20 {
		t.Fatalf("expected all 20 rejected, got %d", rejected)
	}
}
