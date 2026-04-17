package runner

import (
	"context"
	"testing"

	"github.com/user/grpcurl-batch/internal/manifest"
)

func TestPipeline_ContextCancelled_StopsEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPipeline(
		func(_ context.Context, c manifest.Call) (manifest.Call, error) {
			return c, nil
		},
	)
	_, err := p.Run(ctx, baseCall("ctx-test"))
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestPipeline_FullChain_HeadersAndValidation(t *testing.T) {
	p := NewPipeline(
		WithCallValidation(),
		WithHeaderInjection(map[string]string{
			"authorization": "Bearer token",
			"x-request-id": "123",
		}),
	)

	calls := []manifest.Call{
		baseCall("call-1"),
		baseCall("call-2"),
	}

	for _, c := range calls {
		out, err := p.Run(context.Background(), c)
		if err != nil {
			t.Fatalf("call %q failed: %v", c.Name, err)
		}
		if out.Headers["authorization"] != "Bearer token" {
			t.Errorf("call %q missing authorization header", c.Name)
		}
	}
}
