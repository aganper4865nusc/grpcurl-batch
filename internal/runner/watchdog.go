package runner

import (
	"context"
	"sync"
	"time"
)

// WatchdogPolicy monitors long-running calls and cancels them if they exceed a
// stall threshold — distinct from a hard timeout in that it resets on progress.
type WatchdogPolicy struct {
	stallTimeout time.Duration
	mu           sync.Mutex
}

// NewWatchdogPolicy creates a WatchdogPolicy that cancels a call if no
// progress is signalled within stallTimeout.
func NewWatchdogPolicy(stallTimeout time.Duration) *WatchdogPolicy {
	if stallTimeout <= 0 {
		stallTimeout = 10 * time.Second
	}
	return &WatchdogPolicy{stallTimeout: stallTimeout}
}

// Wrap executes fn under watchdog supervision. The returned ProgressFunc must
// be called periodically by the caller to reset the stall timer; if it is not
// called within stallTimeout the context is cancelled.
func (w *WatchdogPolicy) Wrap(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error) {
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTimer(w.stallTimeout)
	defer timer.Stop()

	progress := make(chan struct{}, 1)

	// watchdog goroutine
	go func() {
		for {
			select {
			case <-timer.C:
				cancel()
				return
			case _, ok := <-progress:
				if !ok {
					return
				}
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(w.stallTimeout)
			case <-wctx.Done():
				return
			}
		}
	}()

	// Inject a ping function via context value so callers can signal progress.
	pingCtx := context.WithValue(wctx, watchdogPingKey{}, func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		select {
		case progress <- struct{}{}:
		default:
		}
	})

	res, err := fn(pingCtx)
	close(progress)
	return res, err
}

// PingWatchdog signals progress to a watchdog, if one is present in ctx.
func PingWatchdog(ctx context.Context) {
	if fn, ok := ctx.Value(watchdogPingKey{}).(func()); ok && fn != nil {
		fn()
	}
}

type watchdogPingKey struct{}
