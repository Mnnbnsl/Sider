package persistence

import (
	"os"
	"strconv"
	"time"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/store"
)

var lastSnapshot = time.Now()

// RDBExists reports whether an RDB snapshot file exists.
func RDBExists() bool {
	return fileExists(config.RDBFile)
}

// TakeSnapshot writes the current in-memory store to the RDB snapshot file.
func TakeSnapshot() error {
	fp, err := os.OpenFile(config.RDBFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	store.Range(func(key string, obj *store.Obj) {
		tokens := []string{"SET", key, obj.Value.(string)}
		if obj.ExpiresAt != -1 {
			tokens = append(tokens, "EX", strconv.FormatInt(obj.ExpiresAt, 10))
		}
		if err := writeEntry(fp, tokens); err != nil {
			return
		}
	})

	if err := fp.Sync(); err != nil {
		return err
	}

	lastSnapshot = time.Now()
	return nil
}

// LoadSnapshot loads the RDB snapshot into the in-memory store when it exists.
func LoadSnapshot() error {
	if !RDBExists() {
		return nil
	}
	fp, err := os.Open(config.RDBFile)
	if err != nil {
		return err
	}
	entries, err := readEntries(fp)
	fp.Close()
	if err != nil {
		return err
	}
	for _, tokens := range entries {
		applyTokens(tokens)
	}
	return nil
}

// MaybeSnapshot triggers a snapshot when the configured interval has elapsed.
func MaybeSnapshot() {
	if time.Since(lastSnapshot) < time.Duration(config.SnapshotIntervalMs) {
		return
	}
	if err := TakeSnapshot(); err != nil {
		// best-effort snapshot; errors are logged by the caller side if needed
		_ = err
	}
}
