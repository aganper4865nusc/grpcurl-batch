package manifest

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// RetryConfig defines retry behavior for a gRPC call.
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
}

// Call represents a single gRPC call definition.
type Call struct {
	Name     string            `yaml:"name"`
	Address  string            `yaml:"address"`
	Service  string            `yaml:"service"`
	Method   string            `yaml:"method"`
	Data     string            `yaml:"data"`
	Headers  map[string]string `yaml:"headers"`
	Timeout  time.Duration     `yaml:"timeout"`
	Retry    RetryConfig       `yaml:"retry"`
	Plaintext bool             `yaml:"plaintext"`
}

// Manifest is the top-level structure of the YAML batch manifest.
type Manifest struct {
	Version     string `yaml:"version"`
	Concurrency int    `yaml:"concurrency"`
	Calls       []Call `yaml:"calls"`
}

// Load reads and parses a YAML manifest file from the given path.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest YAML: %w", err)
	}

	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &m, nil
}

// validate checks that the manifest has required fields and sane defaults.
func (m *Manifest) validate() error {
	if len(m.Calls) == 0 {
		return fmt.Errorf("manifest must define at least one call")
	}

	if m.Concurrency <= 0 {
		m.Concurrency = 1
	}

	for i, c := range m.Calls {
		if c.Address == "" {
			return fmt.Errorf("call[%d] %q: address is required", i, c.Name)
		}
		if c.Service == "" {
			return fmt.Errorf("call[%d] %q: service is required", i, c.Name)
		}
		if c.Method == "" {
			return fmt.Errorf("call[%d] %q: method is required", i, c.Name)
		}
		if m.Calls[i].Retry.MaxAttempts <= 0 {
			m.Calls[i].Retry.MaxAttempts = 1
		}
		if m.Calls[i].Timeout == 0 {
			m.Calls[i].Timeout = 30 * time.Second
		}
	}

	return nil
}
