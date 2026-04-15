package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestHookRegistry_FireCallsRegisteredHooks(t *testing.T) {
	reg := NewHookRegistry()
	var called int32
	reg.Register(HookBeforeCall, func(_ context.Context, p HookPayload) {
		atomic.AddInt32(&called, 1)
	})
	reg.Fire(context.Background(), HookPayload{Event: HookBeforeCall, CallName: "svc/Method"})
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected hook to be called once, got %d", called)
	}
}

func TestHookRegistry_MultipleHooksSameEvent(t *testing.T) {
	reg := NewHookRegistry()
	var count int32
	for i := 0; i < 3; i++ {
		reg.Register(HookAfterCall, func(_ context.Context, _ HookPayload) {
			atomic.AddInt32(&count, 1)
		})
	}
	reg.Fire(context.Background(), HookPayload{Event: HookAfterCall})
	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 hooks called, got %d", count)
	}
}

func TestHookRegistry_UnregisteredEvent_NoOp(t *testing.T) {
	reg := NewHookRegistry()
	// Should not panic when no hooks are registered for the event.
	reg.Fire(context.Background(), HookPayload{Event: HookOnRetry, CallName: "svc/Method"})
}

func TestHookRegistry_NilHook_Ignored(t *testing.T) {
	reg := NewHookRegistry()
	reg.Register(HookBeforeCall, nil)
	// Should not panic.
	reg.Fire(context.Background(), HookPayload{Event: HookBeforeCall})
	if len(reg.hooks[HookBeforeCall]) != 0 {
		t.Fatal("nil hook should not be registered")
	}
}

func TestHookRegistry_PanicInHook_Recovered(t *testing.T) {
	reg := NewHookRegistry()
	reg.Register(HookOnFailure, func(_ context.Context, _ HookPayload) {
		panic("intentional panic")
	})
	// Should not propagate the panic.
	reg.Fire(context.Background(), HookPayload{Event: HookOnFailure, CallName: "svc/Method"})
}

func TestHookRegistry_PayloadFieldsPassedThrough(t *testing.T) {
	reg := NewHookRegistry()
	sentErr := errors.New("some error")
	var got HookPayload
	reg.Register(HookOnRetry, func(_ context.Context, p HookPayload) {
		got = p
	})
	want := HookPayload{
		Event:    HookOnRetry,
		CallName: "svc/Retry",
		Attempt:  2,
		Elapsed:  150 * time.Millisecond,
		Err:      sentErr,
	}
	reg.Fire(context.Background(), want)
	if got.CallName != want.CallName || got.Attempt != want.Attempt || !errors.Is(got.Err, sentErr) {
		t.Fatalf("payload mismatch: got %+v", got)
	}
}

func TestDefaultLoggingHook_DoesNotPanic(t *testing.T) {
	hook := DefaultLoggingHook()
	events := []HookEvent{HookBeforeCall, HookAfterCall, HookOnRetry, HookOnFailure}
	for _, ev := range events {
		hook(context.Background(), HookPayload{
			Event:    ev,
			CallName: "svc/Test",
			Attempt:  1,
			Elapsed:  10 * time.Millisecond,
			Err:      errors.New("test error"),
		})
	}
}
