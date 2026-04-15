package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/user/grpcurl-batch/internal/manifest"
)

// Executor defines the interface for executing a single gRPC call.
type Executor interface {
	Execute(ctx context.Context, call manifest.Call) (string, error)
}

// GrpcurlExecutor executes gRPC calls using the grpcurl CLI.
type GrpcurlExecutor struct {
	BinaryPath string
}

// NewGrpcurlExecutor creates a new GrpcurlExecutor.
func NewGrpcurlExecutor(binaryPath string) *GrpcurlExecutor {
	if binaryPath == "" {
		binaryPath = "grpcurl"
	}
	return &GrpcurlExecutor{BinaryPath: binaryPath}
}

// Execute runs grpcurl with the given call parameters.
func (e *GrpcurlExecutor) Execute(ctx context.Context, call manifest.Call) (string, error) {
	args := buildArgs(call)
	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	_ = time.Since(start)

	if err != nil {
		return "", fmt.Errorf("grpcurl error: %w — stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// buildArgs constructs grpcurl CLI arguments from a Call definition.
func buildArgs(call manifest.Call) []string {
	args := []string{"-plaintext"}
	for k, v := range call.Metadata {
		args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
	}
	if call.Data != "" {
		args = append(args, "-d", call.Data)
	}
	args = append(args, call.Address, call.Method)
	return args
}
