package runner

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// CheckpointEntry records the outcome of a single call for resume support.
type CheckpointEntry struct {
	CallID    string    `json:"call_id"`
	Succeeded bool      `json:"succeeded"`
	Timestamp time.Time `json:"timestamp"`
}

// CheckpointStore persists completed call IDs so a batch can be resumed.
type CheckpointStore struct {
	mu      sync.RWMutex
	seen    map[string]CheckpointEntry
	path    string
}

// NewCheckpointStore loads an existing checkpoint file or starts fresh.
func NewCheckpointStore(path string) (*CheckpointStore, error) {
	cs := &CheckpointStore{path: path, seen: make(map[string]CheckpointEntry)}
	if path == "" {
		return cs, nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cs, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []CheckpointEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		cs.seen[e.CallID] = e
	}
	return cs, nil
}

// IsDone reports whether a call has already been completed successfully.
func (cs *CheckpointStore) IsDone(callID string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	e, ok := cs.seen[callID]
	return ok && e.Succeeded
}

// Record marks a call as completed and flushes to disk.
func (cs *CheckpointStore) Record(callID string, succeeded bool) error {
	cs.mu.Lock()
	cs.seen[callID] = CheckpointEntry{CallID: callID, Succeeded: succeeded, Timestamp: time.Now()}
	cs.mu.Unlock()
	return cs.flush()
}

// flush writes current state to the checkpoint file.
func (cs *CheckpointStore) flush() error {
	if cs.path == "" {
		return nil
	}
	cs.mu.RLock()
	entries := make([]CheckpointEntry, 0, len(cs.seen))
	for _, e := range cs.seen {
		entries = append(entries, e)
	}
	cs.mu.RUnlock()
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, data, 0644)
}

// Reset clears all checkpoint data and removes the file.
func (cs *CheckpointStore) Reset() error {
	cs.mu.Lock()
	cs.seen = make(map[string]CheckpointEntry)
	cs.mu.Unlock()
	if cs.path != "" {
		if err := os.Remove(cs.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
