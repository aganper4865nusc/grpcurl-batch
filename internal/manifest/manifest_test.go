package manifest

import (
	"os"
	"testing"
	"time"
)

func writeTempManifest(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "manifest-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidManifest(t *testing.T) {
	yamlContent := `
version: "1"
concurrency: 3
calls:
  - name: get-user
    address: localhost:50051
    service: users.UserService
    method: GetUser
    data: '{"id": "123"}'
    plaintext: true
    timeout: 10s
    retry:
      max_attempts: 3
      delay: 500ms
`
	path := writeTempManifest(t, yamlContent)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Concurrency != 3 {
		t.Errorf("expected concurrency 3, got %d", m.Concurrency)
	}
	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}

	call := m.Calls[0]
	if call.Name != "get-user" {
		t.Errorf("expected name 'get-user', got %q", call.Name)
	}
	if call.Retry.MaxAttempts != 3 {
		t.Errorf("expected max_attempts 3, got %d", call.Retry.MaxAttempts)
	}
	if call.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", call.Timeout)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	yamlContent := `
calls:
  - address: localhost:50051
    service: svc.Svc
    method: Do
`
	path := writeTempManifest(t, yamlContent)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Concurrency != 1 {
		t.Errorf("expected default concurrency 1, got %d", m.Concurrency)
	}
	if m.Calls[0].Retry.MaxAttempts != 1 {
		t.Errorf("expected default max_attempts 1, got %d", m.Calls[0].Retry.MaxAttempts)
	}
	if m.Calls[0].Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", m.Calls[0].Timeout)
	}
}

func TestLoad_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"no calls", `concurrency: 1\ncalls: []`},
		{"missing address", "calls:\n  - service: svc\n    method: M\n"},
		{"missing service", "calls:\n  - address: h:1\n    method: M\n"},
		{"missing method", "calls:\n  - address: h:1\n    service: svc\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempManifest(t, tc.content)
			_, err := Load(path)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/manifest.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
