package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/grpcurl-batch/internal/runner"
)

func makeSpans() []runner.Span {
	now := time.Now()
	return []runner.Span{
		{CallID: "svc/OK", Method: "OK", Start: now, End: now.Add(10 * time.Millisecond), Success: true},
		{CallID: "svc/Fail", Method: "Fail", Start: now, End: now.Add(5 * time.Millisecond), Success: false, Err: errors.New("timeout")},
	}
}

func TestTracingFormatter_Text_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	tf := NewTracingFormatter(&buf, "text")
	if err := tf.Write(makeSpans()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, h := range []string{"CALL", "METHOD", "DURATION", "SUCCESS", "ERROR"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q", h)
		}
	}
}

func TestTracingFormatter_Text_ContainsRows(t *testing.T) {
	var buf bytes.Buffer
	tf := NewTracingFormatter(&buf, "text")
	tf.Write(makeSpans())
	out := buf.String()
	if !strings.Contains(out, "svc/OK") {
		t.Error("missing svc/OK row")
	}
	if !strings.Contains(out, "timeout") {
		t.Error("missing error string")
	}
}

func TestTracingFormatter_JSON_ValidStructure(t *testing.T) {
	var buf bytes.Buffer
	tf := NewTracingFormatter(&buf, "json")
	if err := tf.Write(makeSpans()); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestTracingFormatter_JSON_ErrorField(t *testing.T) {
	var buf bytes.Buffer
	tf := NewTracingFormatter(&buf, "json")
	tf.Write(makeSpans())
	var rows []map[string]interface{}
	json.Unmarshal(buf.Bytes(), &rows)
	if rows[1]["error"] != "timeout" {
		t.Errorf("expected error field 'timeout', got %v", rows[1]["error"])
	}
}

func TestTracingFormatter_NilWriter_UsesStdout(t *testing.T) {
	tf := NewTracingFormatter(nil, "text")
	if tf.w == nil {
		t.Error("expected non-nil writer")
	}
}
