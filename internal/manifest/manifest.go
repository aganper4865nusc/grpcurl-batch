package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Call represents a single gRPC call definition in the manifest.
type Call struct {
	Name    string            `yaml:"name"`
	Service string            `yaml:"service"`
	Method  string            `yaml:"method"`
	Address string            `yaml:"address"`
	Data    string            `yaml:"data"`
	Headers map[string]string `yaml:"headers"`
	Tags    []string          `yaml:"tags"`
	Timeout int               `yaml:"timeout"`
	Retries int               `yaml:"retries"`
}

// Manifest holds the full batch manifest configuration.
type Manifest struct {
	Version     string `yaml:"version"`
	Concurrency int    `yaml:"concurrency"`
	Calls       []Call `yaml:"calls"`
}

// Load reads and parses a manifest YAML file from the given path.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	applyDefaults(&m)
	if err := validate(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func applyDefaults(m *Manifest) {
	if m.Concurrency <= 0 {
		m.Concurrency = 1
	}
	for i := range m.Calls {
		if m.Calls[i].Timeout == 0 {
			m.Calls[i].Timeout = 30
		}
		if m.Calls[i].Retries == 0 {
			m.Calls[i].Retries = 1
		}
	}
}

func validate(m *Manifest) error {
	if len(m.Calls) == 0 {
		return fmt.Errorf("manifest must contain at least one call")
	}
	for i, c := range m.Calls {
		if c.Service == "" {
			return fmt.Errorf("call[%d] %q: service is required", i, c.Name)
		}
		if c.Method == "" {
			return fmt.Errorf("call[%d] %q: method is required", i, c.Name)
		}
		if c.Address == "" {
			return fmt.Errorf("call[%d] %q: address is required", i, c.Name)
		}
	}
	return nil
}
