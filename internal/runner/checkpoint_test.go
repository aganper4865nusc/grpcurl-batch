package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func tmpCheckpointPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestCheckpointStore_NewFile_Empty(t *testing.T) {
	cs, err := NewCheckpointStore(tmpCheckpointPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cs.IsDone("call-1") {
		t.Error("expected call-1 to not be done")
	}
}

func TestCheckpointStore_RecordAndIsDone(t *testing.T) {
	cs, _ := NewCheckpointStore(tmpCheckpointPath(t))
	if err := cs.Record("call-1", true); err != nil {
		t.Fatalf("record error: %v", err)
	}
	if !cs.IsDone("call-1") {
		t.Error("expected call-1 to be done")
	}
}

func TestCheckpointStore_FailedCall_NotDone(t *testing.T) {
	cs, _ := NewCheckpointStore(tmpCheckpointPath(t))
	_ = cs.Record("call-2", false)
	if cs.IsDone("call-2") {
		t.Error("failed call should not be marked done")
	}
}

func TestCheckpointStore_PersistsAcrossReload(t *testing.T) {
	path := tmpCheckpointPath(t)
	cs, _ := NewCheckpointStore(path)
	_ = cs.Record("call-3", true)

	cs2, err := NewCheckpointStore(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if !cs2.IsDone("call-3") {
		t.Error("expected call-3 to persist across reload")
	}
}

func TestCheckpointStore_Reset_ClearsState(t *testing.T) {
	path := tmpCheckpointPath(t)
	cs, _ := NewCheckpointStore(path)
	_ = cs.Record("call-4", true)
	if err := cs.Reset(); err != nil {
		t.Fatalf("reset error: %v", err)
	}
	if cs.IsDone("call-4") {
		t.Error("expected state to be cleared after reset")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected checkpoint file to be removed")
	}
}

func TestCheckpointStore_EmptyPath_NoFile(t *testing.T) {
	cs, err := NewCheckpointStore("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cs.Record("call-5", true); err != nil {
		t.Fatalf("record with no path should not error: %v", err)
	}
	if !cs.IsDone("call-5") {
		t.Error("in-memory record should still work")
	}
}
