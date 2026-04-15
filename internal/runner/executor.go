package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/your-org/grpcurl-batch/internal/manifest"
)

// GrpcurlExecutor runs grpcurl as a subprocess.
type GrpcurlExecutor struct {
	BinaryPath string
	Timeout    time.Duration
}

// NewGrpcurlExecutor creates an executor using the grpcurl binary.
func NewGrpcurlExecutor(binaryPath string, timeout time.Duration) *GrpcurlExecutor {
	if binaryPath == "" {
		binaryPath = "grpcurl"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &GrpcurlExecutor{BinaryPath: binaryPath, Timeout: timeout}
}

// Execute builds and runs a grpcurl command for the given call.
func (g *GrpcurlExecutor) Execute(ctx context.Context, call manifest.Call) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, g.Timeout)
	defer cancel()

	args := g.buildArgs(call)
	cmd := exec.CommandContext(ctx, g.BinaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("grpcurl failed for %q: %w — stderr: %s", call.Name, err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (g *GrpcurlExecutor) buildArgs(call manifest.Call) []string {
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
