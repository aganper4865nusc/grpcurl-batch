package config

import (
	"os"
	"testing"
	"time"
)

func writeTempFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "manifest-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.OutputFormat != "text" {
		t.Errorf("expected output format \"text\", got %q", cfg.OutputFormat)
	}
	if cfg.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", cfg.Concurrency)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %s", cfg.Timeout)
	}
}

func TestValidate_MissingManifest(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing manifest path")
	}
}

func TestValidate_FileNotFound(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ManifestPath = "/nonexistent/path/manifest.yaml"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidate_InvalidOutputFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ManifestPath = writeTempFile(t)
	cfg.OutputFormat = "xml"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid output format")
	}
}

func TestValidate_InvalidConcurrency(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ManifestPath = writeTempFile(t)
	cfg.Concurrency = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for concurrency < 1")
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ManifestPath = writeTempFile(t)
	cfg.Timeout = -1 * time.Second
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for non-positive timeout")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ManifestPath = writeTempFile(t)
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}
