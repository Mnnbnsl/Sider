package command

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/resp"
	"github.com/mnnbnsl/sider/internal/store"
)

func EvalTTL(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'TTL' command"), false)
	}
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return resp.RESP_TTL_NOEXIST
	}

	if obj.ExpiresAt == -1 {
		return resp.RESP_TTL_NOEXPIRE
	}

	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		return resp.RESP_TTL_NOEXIST
	}

	return resp.Encode(int64(durationMs/1000), false)
}

func EvalDEL(args []string) []byte {
	if len(args) < 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'DEL' command"), false)
	}

	deleted := store.Del(args)
	if len(args) > 0 {
		if err := persistence.AppendDel(args...); err != nil {
			return resp.Encode(fmt.Errorf("ERR failed to persist DEL: %w", err), false)
		}
	}
	return resp.Encode(deleted, false)
}

func EvalEXPIRE(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'EXPIRE' command"), false)
	}

	key := args[0]
	durationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	durationMs := durationSec * 1000

	result := store.Expire(key, durationMs)
	if result > 0 {
		obj := store.Get(key)
		if obj != nil {
			if err := persistence.AppendSet(key, obj.Value.(string), obj.ExpiresAt); err != nil {
				return resp.Encode(fmt.Errorf("ERR failed to persist EXPIRE: %w", err), false)
			}
		}
	}
	return resp.Encode(result, false)
}
