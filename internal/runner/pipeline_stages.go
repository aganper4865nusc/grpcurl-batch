package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// WithAddressOverride returns a PipelineStage that replaces the call's address.
func WithAddressOverride(address string) PipelineStage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if address != "" {
			call.Address = address
		}
		return call, nil
	}
}

// WithTagFilter returns a PipelineStage that rejects calls not matching the tag.
func WithTagFilter(tag string) PipelineStage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if tag == "" {
			return call, nil
		}
		if !hasAnyTag(call, []string{tag}) {
			return call, fmt.Errorf("call %q does not match tag %q", call.Method, tag)
		}
		return call, nil
	}
}

// WithMaxRetries returns a PipelineStage that caps the call's retry count.
func WithMaxRetries(max int) PipelineStage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if call.Retries > max {
			call.Retries = max
		}
		return call, nil
	}
}

// WithDeadlineGuard returns a PipelineStage that rejects calls when the batch deadline is exceeded.
func WithDeadlineGuard(policy *DeadlinePolicy) PipelineStage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if policy.Exceeded() {
			return call, fmt.Errorf("batch deadline exceeded, skipping call %q", call.Method)
		}
		return call, nil
	}
}

// WithTimestampHeader returns a PipelineStage that injects an X-Timestamp header.
func WithTimestampHeader() PipelineStage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if call.Headers == nil {
			call.Headers = map[string]string{}
		}
		call.Headers["x-timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
		return call, nil
	}
}
