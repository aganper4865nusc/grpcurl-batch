package runner

import (
	"context"
	"log"
	"time"
)

// HookEvent represents the lifecycle stage of a call.
type HookEvent string

const (
	HookBeforeCall HookEvent = "before_call"
	HookAfterCall  HookEvent = "after_call"
	HookOnRetry    HookEvent = "on_retry"
	HookOnFailure  HookEvent = "on_failure"
)

// HookPayload carries contextual data passed to each hook handler.
type HookPayload struct {
	Event     HookEvent
	CallName  string
	Attempt   int
	Elapsed   time.Duration
	Err       error
}

// HookFunc is a function invoked at a lifecycle event.
type HookFunc func(ctx context.Context, payload HookPayload)

// HookRegistry holds registered hooks for each lifecycle event.
type HookRegistry struct {
	hooks map[HookEvent][]HookFunc
}

// NewHookRegistry creates an empty HookRegistry.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookEvent][]HookFunc),
	}
}

// Register adds a HookFunc for the given event.
func (r *HookRegistry) Register(event HookEvent, fn HookFunc) {
	if fn == nil {
		return
	}
	r.hooks[event] = append(r.hooks[event], fn)
}

// Fire invokes all hooks registered for the given event.
// Hooks are called synchronously in registration order.
// Panics inside hooks are recovered and logged.
func (r *HookRegistry) Fire(ctx context.Context, payload HookPayload) {
	for _, fn := range r.hooks[payload.Event] {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("[hooks] panic in %s hook for call %q: %v", payload.Event, payload.CallName, rec)
				}
			}()
			fn(ctx, payload)
		}()
	}
}

// DefaultLoggingHook returns a HookFunc that logs each event.
func DefaultLoggingHook() HookFunc {
	return func(_ context.Context, p HookPayload) {
		switch p.Event {
		case HookBeforeCall:
			log.Printf("[hooks] starting call %q (attempt %d)", p.CallName, p.Attempt)
		case HookAfterCall:
			log.Printf("[hooks] finished call %q in %s (attempt %d)", p.CallName, p.Elapsed, p.Attempt)
		case HookOnRetry:
			log.Printf("[hooks] retrying call %q after error: %v (attempt %d)", p.CallName, p.Err, p.Attempt)
		case HookOnFailure:
			log.Printf("[hooks] call %q failed permanently: %v", p.CallName, p.Err)
		}
	}
}
