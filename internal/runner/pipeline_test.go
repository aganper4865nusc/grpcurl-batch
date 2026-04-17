package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/user/grpcurl-batch/internal/manifest"
)

func baseCall(name string) manifest.Call {
	return manifest.Call{Name: name, Service: "svc", Method: "Method", Address: "localhost:50051"}
}

func TestPipeline_EmptyStages_ReturnsCall(t *testing.T) {
	p := NewPipeline()
	call := baseCall("a")
	out, err := p.Run(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "a" {
		t.Errorf("expected name 'a', got %q", out.Name)
	}
}

func TestPipeline_StagesAppliedInOrder(t *testing.T) {
	var order []int
	mk := func(n int) Stage {
		return func(_ context.Context, c manifest.Call) (manifest.Call, error) {
			order = append(order, n)
			return c, nil
		}
	}
	p := NewPipeline(mk(1), mk(2), mk(3))
	_, _ = p.Run(context.Background(), baseCall("b"))
	for i, v := range order {
		if v != i+1 {
			t.Errorf("expected stage %d at index %d, got %d", i+1, i, v)
		}
	}
}

func TestPipeline_StageError_StopsChain(t *testing.T) {
	called := false
	p := NewPipeline(
		func(_ context.Context, c manifest.Call) (manifest.Call, error) {
			return c, errors.New("boom")
		},
		func(_ context.Context, c manifest.Call) (manifest.Call, error) {
			called = true
			return c, nil
		},
	)
	_, err := p.Run(context.Background(), baseCall("c"))
	if err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Error("second stage should not have been called")
	}
}

func TestWithHeaderInjection_InjectsHeaders(t *testing.T) {
	stage := WithHeaderInjection(map[string]string{"x-token": "abc"})
	call := baseCall("d")
	out, err := stage(context.Background(), call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Headers["x-token"] != "abc" {
		t.Errorf("expected header x-token=abc, got %q", out.Headers["x-token"])
	}
}

func TestWithHeaderInjection_DoesNotOverwrite(t *testing.T) {
	stage := WithHeaderInjection(map[string]string{"x-token": "new"})
	call := baseCall("e")
	call.Headers = map[string]string{"x-token": "original"}
	out, _ := stage(context.Background(), call)
	if out.Headers["x-token"] != "original" {
		t.Errorf("expected original header to be preserved")
	}
}

func TestWithCallValidation_MissingService(t *testing.T) {
	stage := WithCallValidation()
	call := manifest.Call{Name: "f", Method: "M"}
	_, err := stage(context.Background(), call)
	if err == nil {
		t.Fatal("expected error for missing service")
	}
}

func TestWithCallValidation_Valid(t *testing.T) {
	stage := WithCallValidation()
	_, err := stage(context.Background(), baseCall("g"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
