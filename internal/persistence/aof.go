package persistence

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/resp"
	"github.com/mnnbnsl/sider/internal/store"
)

var (
	aofFile     *os.File
	lastRewrite time.Time
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getAppendFile() (*os.File, error) {
	if aofFile != nil {
		return aofFile, nil
	}
	f, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	aofFile = f
	return f, nil
}

func writeEntry(fp *os.File, tokens []string) error {
	_, err := fp.Write(resp.Encode(tokens, false))
	return err
}

func applyTokens(tokens []string) {
	if len(tokens) == 0 {
		return
	}

	switch tokens[0] {
	case "SET":
		if len(tokens) < 3 {
			return
		}
		key := tokens[1]
		value := tokens[2]
		var exMs int64 = -1
		if len(tokens) >= 5 && tokens[3] == "EX" {
			if parsed, err := strconv.ParseInt(tokens[4], 10, 64); err == nil {
				exMs = parsed
			}
		}
		oType, oEnc := store.DeduceTypeEncoding(value)
		store.Put(key, &store.Obj{
			TypeEncoding: oType | oEnc,
			Value:        value,
			ExpiresAt:    exMs,
		})
	case "DEL":
		if len(tokens) > 1 {
			store.Del(tokens[1:])
		}
	}
}

func readEntries(fp *os.File) ([][]string, error) {
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	values, err := resp.Decode(data)
	if err != nil {
		return nil, err
	}
	var entries [][]string
	for _, v := range values {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		tokens := make([]string, len(arr))
		valid := true
		for i, item := range arr {
			s, ok := item.(string)
			if !ok {
				valid = false
				break
			}
			tokens[i] = s
		}
		if valid {
			entries = append(entries, tokens)
		}
	}
	return entries, nil
}

// LoadAOF replays the AOF file into the in-memory store when it exists and
// prepares the append handle for future writes. If a checkpoint exists, the
// part of the AOF that was already reflected in the RDB snapshot is skipped.
func LoadAOF() error {
	if fileExists(config.AOFFile) {
		skipFrom := int64(0)
		if cp, ok, err := readCheckpoint(); err != nil {
			return err
		} else if ok && RDBExists() && cp.AofSize <= AOFSize() {
			skipFrom = cp.AofSize
		}

		fp, err := os.Open(config.AOFFile)
		if err != nil {
			return err
		}
		if skipFrom > 0 {
			if _, err := fp.Seek(skipFrom, io.SeekStart); err != nil {
				fp.Close()
				return err
			}
		}
		entries, err := readEntries(fp)
		fp.Close()
		if err != nil {
			return err
		}
		for _, tokens := range entries {
			applyTokens(tokens)
		}
	}

	f, err := getAppendFile()
	if err != nil {
		return err
	}
	aofFile = f
	lastRewrite = time.Now()
	return nil
}

// AppendSet appends a SET entry (with optional EX) to the AOF.
func AppendSet(key string, value string, exMs int64) error {
	tokens := []string{"SET", key, value}
	if exMs != -1 {
		tokens = append(tokens, "EX", strconv.FormatInt(exMs, 10))
	}

	f, err := getAppendFile()
	if err != nil {
		return err
	}
	if _, err := f.Write(resp.Encode(tokens, false)); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	if AOFSize() > config.AOFMaxSizeBytes {
		return RewriteAOF()
	}
	return nil
}

// AppendDel appends a DEL entry to the AOF.
func AppendDel(keys ...string) error {
	tokens := []string{"DEL"}
	tokens = append(tokens, keys...)

	f, err := getAppendFile()
	if err != nil {
		return err
	}
	if _, err := f.Write(resp.Encode(tokens, false)); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	if AOFSize() > config.AOFMaxSizeBytes {
		return RewriteAOF()
	}
	return nil
}

// RewriteAOF truncates the AOF file and rewrites it from the in-memory store.
func RewriteAOF() error {
	if aofFile != nil {
		aofFile.Close()
		aofFile = nil
	}

	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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

	lastRewrite = time.Now()
	if err := clearCheckpoint(); err != nil {
		return err
	}
	return nil
}

// AOFSize returns the current size of the AOF file in bytes.
func AOFSize() int64 {
	info, err := os.Stat(config.AOFFile)
	if err != nil {
		return 0
	}
	return info.Size()
}

// MaybeRewriteAOF triggers an AOF rewrite when the file has grown past the
// configured size limit or when the rewrite interval has elapsed.
func MaybeRewriteAOF() {
	if time.Since(lastRewrite) < time.Duration(config.AOFRewriteIntervalMs) {
		return
	}
	if AOFSize() > config.AOFMaxSizeBytes {
		_ = RewriteAOF()
	}
}
