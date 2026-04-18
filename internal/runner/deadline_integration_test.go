package runner

import (
	"context"
	"testing"
	"time"
)

func TestDeadlinePolicy_ContextExpires(t *testing.T) {
	p := NewDeadlineFromDuration(50 * time.Millisecond)
	ctx, cancel := p.Wrap(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Error("context should have expired by now")
	}
}

func TestDeadlinePolicy_ExceededAfterExpiry(t *testing.T) {
	p := NewDeadlineFromDuration(30 * time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if !p.Exceeded() {
		t.Error("expected deadline to be exceeded after sleep")
	}
}

func TestDeadlinePolicy_WrapRespectsParentCancellation(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	p := NewDeadlineFromDuration(5 * time.Second)
	ctx, cancel := p.Wrap(parent)
	defer cancel()

	parentCancel()
	select {
	case <-ctx.Done():
		// expected: parent cancellation propagates
	case <-time.After(100 * time.Millisecond):
		t.Error("expected context to be cancelled with parent")
	}
}
