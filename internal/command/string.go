package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mnnbnsl/sider/internal/persistence"
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
	oType, oEnc := store.DeduceTypeEncoding(value)

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
	obj := store.NewObj(value, exDurationMs, oType, oEnc)
	store.Put(key, obj)
	if err := persistence.AppendSet(key, value, obj.ExpiresAt); err != nil {
		return resp.Encode(fmt.Errorf("ERR failed to persist SET: %w", err), false)
	}
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

func EvalINCR(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'INCR' command"), false)
	}
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		obj = store.NewObj("0", -1, store.OBJ_TYPE_STRING, store.OBJ_ENCODING_INT)
		store.Put(key, obj)
	}

	if err := store.AssertType(obj.TypeEncoding, store.OBJ_TYPE_STRING); err != nil {
		return resp.Encode(err, false)
	}

	if err := store.AssertEncoding(obj.TypeEncoding, store.OBJ_ENCODING_INT); err != nil {
		return resp.Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)
	if err := persistence.AppendSet(key, obj.Value.(string), obj.ExpiresAt); err != nil {
		return resp.Encode(fmt.Errorf("ERR failed to persist INCR: %w", err), false)
	}
	return resp.Encode(i, false)
}
