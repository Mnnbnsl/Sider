package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mnnbnsl/sider/config"
)

// Checkpoint records the boundary between the RDB snapshot and the AOF log.
// AofSize is the size of the AOF file (in bytes) that was already reflected in
// the RDB snapshot when the checkpoint was persisted. On startup, the leading
// part of the AOF covered by this boundary is considered redundant (already
// present in the RDB snapshot) and is skipped; only the remaining tail has to
// be replayed.
type Checkpoint struct {
	AofSize    int64     `json:"aofSize"`
	SnapshotAt time.Time `json:"snapshotAt"`
}

// readCheckpoint reads the persisted checkpoint. ok reports whether a valid
// checkpoint exists; it is absent after the first boot and after an AOF
// rewrite, since the rewritten AOF is self-contained again.
func readCheckpoint() (Checkpoint, bool, error) {
	var cp Checkpoint
	data, err := os.ReadFile(config.CheckpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cp, false, nil
		}
		return cp, false, err
	}
	if err := json.Unmarshal(data, &cp); err != nil {
		return cp, false, err
	}
	return cp, true, nil
}

// writeCheckpoint persists the checkpoint atomically (temp file + rename).
func writeCheckpoint(aofSize int64) error {
	cp := Checkpoint{
		AofSize:    aofSize,
		SnapshotAt: time.Now(),
	}
	data, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	tmpPath := config.CheckpointFile + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, config.CheckpointFile); err != nil {
		return err
	}
	if dir, err := os.Open(filepath.Dir(config.CheckpointFile)); err == nil {
		defer dir.Close()
		_ = dir.Sync()
	}
	return nil
}

// clearCheckpoint removes the checkpoint, e.g. after an AOF rewrite which made
// the rewritten AOF self-contained again.
func clearCheckpoint() error {
	if err := os.Remove(config.CheckpointFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
