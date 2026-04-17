package runner

import (
	"context"
	"fmt"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// Stage represents a named pipeline stage that transforms a call before execution.
type Stage func(ctx context.Context, call manifest.Call) (manifest.Call, error)

// Pipeline chains multiple stages applied sequentially to each call.
type Pipeline struct {
	stages []Stage
}

// NewPipeline creates a Pipeline with the given stages.
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Run applies all stages to the call in order, returning the final call or the
// first error encountered.
func (p *Pipeline) Run(ctx context.Context, call manifest.Call) (manifest.Call, error) {
	var err error
	for i, stage := range p.stages {
		call, err = stage(ctx, call)
		if err != nil {
			return call, fmt.Errorf("pipeline stage %d: %w", i, err)
		}
		if ctx.Err() != nil {
			return call, ctx.Err()
		}
	}
	return call, nil
}

// Len returns the number of stages in the pipeline.
func (p *Pipeline) Len() int { return len(p.stages) }

// WithHeaderInjection returns a Stage that injects static headers into a call.
func WithHeaderInjection(headers map[string]string) Stage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if call.Headers == nil {
			call.Headers = make(map[string]string)
		}
		for k, v := range headers {
			if _, exists := call.Headers[k]; !exists {
				call.Headers[k] = v
			}
		}
		return call, nil
	}
}

// WithCallValidation returns a Stage that rejects calls missing required fields.
func WithCallValidation() Stage {
	return func(_ context.Context, call manifest.Call) (manifest.Call, error) {
		if call.Service == "" {
			return call, fmt.Errorf("call %q: service is required", call.Name)
		}
		if call.Method == "" {
			return call, fmt.Errorf("call %q: method is required", call.Name)
		}
		return call, nil
	}
}
