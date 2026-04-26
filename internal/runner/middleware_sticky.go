package runner

import (
	"context"
	"fmt"
)

// StickyKeyFunc derives a routing key from a Call. Returning an empty string
// disables sticky routing for that call.
type StickyKeyFunc func(c Call) string

// CallServiceMethodKey is a StickyKeyFunc that combines the service and method
// as the sticky key, useful for per-endpoint affinity.
func CallServiceMethodKey(c Call) string {
	return fmt.Sprintf("%s/%s", c.Service, c.Method)
}

// WithSticky returns middleware that applies session-affinity routing via
// StickyPolicy. The keyFn derives the routing key per call; if keyFn is nil
// or returns an empty string the call is forwarded unchanged.
func WithSticky(s *StickyPolicy, keyFn StickyKeyFunc) Middleware {
	return func(next CallFunc) CallFunc {
		return func(ctx context.Context, c Call) (Response, error) {
			if s == nil || keyFn == nil {
				return next(ctx, c)
			}
			key := keyFn(c)
			if key == "" {
				return next(ctx, c)
			}
			return s.Wrap(key, next)(ctx, c)
		}
	}
}

// StickyStatus returns a snapshot of all current key→address bindings.
func StickyStatus(s *StickyPolicy) map[string]string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.routes))
	for k, v := range s.routes {
		out[k] = v
	}
	return out
}
