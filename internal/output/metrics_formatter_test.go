package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/grpcurl-batch/internal/runner"
)

func makeSnap() runner.Metrics {
	m := runner.NewMetrics()
	m.RecordSuccess(10 * time.Millisecond)
	m.RecordSuccess(20 * time.Millisecond)
	m.RecordFailure(5 * time.Millisecond)
	m.RecordRetry()
	return m.Snapshot()
}

func TestMetricsFormatter_Text_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	f := NewMetricsFormatter(&buf, "text")
	snap := makeSnap()
	_ = f.Write(snap)
	out := buf.String()
	for _, h := range []string{"Total Calls", "Succeeded", "Failed", "Retried"} {
		if !strings.Contains(out, h) {
			t.Errorf("expected %q in output", h)
		}
	}
}

func TestMetricsFormatter_Text_Values(t *testing.T) {
	var buf bytes.Buffer
	f := NewMetricsFormatter(&buf, "text")
	snap := makeSnap()
	_ = f.Write(snap)
	out := buf.String()
	if !strings.Contains(out, "3") {
		t.Error("expected total calls count in output")
	}
}

func TestMetricsFormatter_JSON_ValidStructure(t *testing.T) {
	var buf bytes.Buffer
	f := NewMetricsFormatter(&buf, "json")
	snap := makeSnap()
	_ = f.Write(snap)
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, key := range []string{"total_calls", "succeeded", "failed", "retried"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}

func TestMetricsFormatter_JSON_Counts(t *testing.T) {
	var buf bytes.Buffer
	f := NewMetricsFormatter(&buf, "json")
	snap := makeSnap()
	_ = f.Write(snap)
	var result map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &result)
	if result["total_calls"].(float64) != 3 {
		t.Errorf("expected total_calls=3, got %v", result["total_calls"])
	}
	if result["retried"].(float64) != 1 {
		t.Errorf("expected retried=1, got %v", result["retried"])
	}
}

func TestMetricsFormatter_NilWriter_UsesStdout(t *testing.T) {
	f := NewMetricsFormatter(nil, "text")
	if f.w == nil {
		t.Error("expected non-nil writer")
	}
}
