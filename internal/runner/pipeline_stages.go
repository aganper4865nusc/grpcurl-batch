package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// WithAddressOverride returns a Stage that replaces the call address if override
// is non-empty.
func WithAddressOverride(address string) Stage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if address != "" {
			call.Address = address
		}
		return call, nil
	}
}

// WithTagFilter returns a Stage that rejects calls not matching any of the
// required tags. If requiredTags is empty all calls are accepted.
func WithTagFilter(requiredTags []string) Stage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if len(requiredTags) == 0 {
			return call, nil
		}
		for _, rt := range requiredTags {
			for _, ct := range call.Tags {
				if strings.EqualFold(ct, rt) {
					return call, nil
				}
			}
		}
		return call, fmt.Errorf("call %q skipped: no matching tag in %v", call.Name, requiredTags)
	}
}

// WithMaxRetries returns a Stage that clamps the retry count to the given max.
func WithMaxRetries(max int) Stage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if max <= 0 {
			return call, nil
		}
		if call.Retries > max {
			call.Retries = max
		}
		return call, nil
	}
}
