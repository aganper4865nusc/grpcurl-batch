package output_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/grpcurl-batch/internal/output"
	"github.com/grpcurl-batch/internal/runner"
)

func makeResults() []runner.Result {
	return []runner.Result{
		{
			CallName: "SayHello",
			Success:  true,
			Attempts: 1,
			Duration: 120 * time.Millisecond,
			Err:      nil,
		},
		{
			CallName: "GetUser",
			Success:  false,
			Attempts: 3,
			Duration: 500 * time.Millisecond,
			Err:      errors.New("deadline exceeded"),
		},
	}
}

func TestFormatter_WriteText_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf)
	if err := f.WriteResults(makeResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, hdr := range []string{"CALL", "STATUS", "ATTEMPTS", "DURATION", "ERROR"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestFormatter_WriteText_ContainsResults(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf)
	_ = f.WriteResults(makeResults())
	out := buf.String()
	if !strings.Contains(out, "SayHello") || !strings.Contains(out, "OK") {
		t.Error("expected SayHello OK in text output")
	}
	if !strings.Contains(out, "GetUser") || !strings.Contains(out, "FAIL") {
		t.Error("expected GetUser FAIL in text output")
	}
	if !strings.Contains(out, "deadline exceeded") {
		t.Error("expected error message in text output")
	}
}

func TestFormatter_WriteJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatJSON, &buf)
	if err := f.WriteResults(makeResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var results []runner.Result
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFormatter_WriteJSON_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatJSON, &buf)
	if err := f.WriteResults([]runner.Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("expected empty JSON array, got %q", buf.String())
	}
}
