package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds the global CLI configuration derived from flags and environment.
type Config struct {
	ManifestPath string
	OutputFormat string // "text" or "json"
	Concurrency  int
	Timeout      time.Duration
	Verbose      bool
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		OutputFormat: "text",
		Concurrency:  4,
		Timeout:      30 * time.Second,
		Verbose:      false,
	}
}

// Validate checks that required fields are present and values are in range.
func (c *Config) Validate() error {
	if c.ManifestPath == "" {
		return fmt.Errorf("manifest path must not be empty")
	}
	if _, err := os.Stat(c.ManifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest file not found: %s", c.ManifestPath)
	}
	if c.OutputFormat != "text" && c.OutputFormat != "json" {
		return fmt.Errorf("output format must be \"text\" or \"json\", got %q", c.OutputFormat)
	}
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be >= 1, got %d", c.Concurrency)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %s", c.Timeout)
	}
	return nil
}
