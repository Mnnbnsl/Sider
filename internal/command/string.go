package command

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mnnbnsl/sider/internal/resp"
	"github.com/mnnbnsl/sider/internal/store"
)

func EvalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'PING' command"), false)
	}

	if len(args) == 0 {
		b = resp.RESP_PONG
	} else {
		b = resp.Encode(args[0], false)
	}

	return b
}

func EvalSET(args []string) []byte {
	if len(args) <= 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'SET' command"), false)
	}
	var exDurationMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch strings.ToLower(args[i]) {
		case "ex":
			i++
			if i == len(args) {
				return resp.Encode(errors.New("ERR syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			exDurationMs = 1000 * exDurationSec
		default:
			return resp.Encode(errors.New("ERR syntax error"), false)
		}
	}
	obj := store.NewObj(value, exDurationMs)
	store.Put(key, obj)
	return resp.RESP_OK
}

func EvalGET(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'GET' command"), false)
	}
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return resp.RESP_NIL
	}

	return resp.Encode(obj.Value, false)
}
