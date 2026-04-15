package manifest

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Call represents a single gRPC call entry in the manifest.
type Call struct {
	Service  string            `yaml:"service"`
	Method   string            `yaml:"method"`
	Address  string            `yaml:"address"`
	Data     string            `yaml:"data"`
	Tags     []string          `yaml:"tags"`
	Headers  map[string]string `yaml:"headers"`
	Timeout  time.Duration     `yaml:"timeout"`
	Retries  int               `yaml:"retries"`
	Insecure bool              `yaml:"insecure"`
}

// Manifest holds the full set of calls and global defaults.
type Manifest struct {
	Defaults CallDefaults `yaml:"defaults"`
	Calls    []Call       `yaml:"calls"`
}

// CallDefaults provides fallback values applied to every call.
type CallDefaults struct {
	Address  string        `yaml:"address"`
	Timeout  time.Duration `yaml:"timeout"`
	Retries  int           `yaml:"retries"`
	Insecure bool          `yaml:"insecure"`
}

// Load reads a YAML manifest from path, applies defaults, and validates it.
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

// applyDefaults fills in missing per-call fields from manifest-level defaults.
func applyDefaults(m *Manifest) {
	for i := range m.Calls {
		c := &m.Calls[i]
		if c.Address == "" {
			c.Address = m.Defaults.Address
		}
		if c.Timeout == 0 {
			c.Timeout = m.Defaults.Timeout
		}
		if c.Retries == 0 {
			c.Retries = m.Defaults.Retries
		}
		if !c.Insecure {
			c.Insecure = m.Defaults.Insecure
		}
	}
}

// validate checks that all required fields are present.
func validate(m *Manifest) error {
	if len(m.Calls) == 0 {
		return errors.New("manifest must contain at least one call")
	}
	for i, c := range m.Calls {
		if c.Service == "" {
			return fmt.Errorf("call[%d]: service is required", i)
		}
		if c.Method == "" {
			return fmt.Errorf("call[%d]: method is required", i)
		}
		if c.Address == "" {
			return fmt.Errorf("call[%d]: address is required (set per-call or in defaults)", i)
		}
	}
	return nil
}
