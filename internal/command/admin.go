package command

import (
	"fmt"
	"strings"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/resp"
	"github.com/mnnbnsl/sider/internal/store"
)

func EvalBGREWRITEAOF(args []string) []byte {
	if err := persistence.RewriteAOF(); err != nil {
		return resp.Encode(fmt.Errorf("ERR failed to rewrite AOF: %w", err), false)
	}
	return resp.RESP_OK
}

func EvalINFO(args []string) []byte {
	if len(args) > 1 {
		return resp.Encode(fmt.Errorf("ERR wrong number of arguments for 'INFO' command"), false)
	}

	stats := store.GetStats()
	info := strings.Builder{}
	info.WriteString("# Keyspace\r\n")
	info.WriteString(fmt.Sprintf("db0:keys=%d,expires=%d\r\n", stats.KeyCount, stats.ExpiringKeys))
	info.WriteString("# Memory\r\n")
	info.WriteString(fmt.Sprintf("used_keys:%d\r\n", stats.KeyCount))
	info.WriteString(fmt.Sprintf("max_keys:%d\r\n", config.KeysLimit))
	info.WriteString(fmt.Sprintf("eviction_policy:%s\r\n", config.EvictionStrategy))

	return resp.Encode(info.String(), false)
}
