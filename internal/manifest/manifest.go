package manifest

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Call represents a single gRPC call definition.
type Call struct {
	Name       string            `yaml:"name"`
	Address    string            `yaml:"address"`
	Method     string            `yaml:"method"`
	Data       string            `yaml:"data"`
	Metadata   map[string]string `yaml:"metadata"`
	Retries    int               `yaml:"retries"`
	RetryDelay time.Duration     `yaml:"retry_delay"`
	Timeout    time.Duration     `yaml:"timeout"`
}

// Manifest is the top-level structure of a batch manifest file.
type Manifest struct {
	Version     string `yaml:"version"`
	Concurrency int    `yaml:"concurrency"`
	Calls       []Call `yaml:"calls"`
}

const (
	defaultConcurrency = 5
	defaultRetries     = 0
	defaultRetryDelay  = 500 * time.Millisecond
	defaultTimeout     = 30 * time.Second
)

// Load reads and parses a manifest YAML file, applying defaults.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %q: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest %q: %w", path, err)
	}

	if err := validate(&m); err != nil {
		return nil, err
	}

	applyDefaults(&m)
	return &m, nil
}

func validate(m *Manifest) error {
	for i, c := range m.Calls {
		if c.Name == "" {
			return fmt.Errorf("call[%d]: missing required field 'name'", i)
		}
		if c.Address == "" {
			return fmt.Errorf("call %q: missing required field 'address'", c.Name)
		}
		if c.Method == "" {
			return fmt.Errorf("call %q: missing required field 'method'", c.Name)
		}
	}
	return nil
}

func applyDefaults(m *Manifest) {
	if m.Concurrency <= 0 {
		m.Concurrency = defaultConcurrency
	}
	for i := range m.Calls {
		if m.Calls[i].RetryDelay == 0 {
			m.Calls[i].RetryDelay = defaultRetryDelay
		}
		if m.Calls[i].Timeout == 0 {
			m.Calls[i].Timeout = defaultTimeout
		}
	}
}
