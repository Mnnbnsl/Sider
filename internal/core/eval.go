package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mnnbnsl/sider/internal/resp"
)

var (
	// RESP_NIL is the null bulk string reply.
	RESP_NIL []byte = []byte("$-1\r\n")
	// RESP_OK is the simple string reply for successful commands.
	RESP_OK []byte = []byte("+OK\r\n")
	// RESP_PONG is the reply returned by PING without arguments.
	RESP_PONG []byte = []byte("+PONG\r\n")
	// RESP_EMPTY_ARRAY is an empty array reply for unsupported discovery commands.
	RESP_EMPTY_ARRAY []byte = []byte("*0\r\n")
	// RESP_TTL_NOEXPIRE means the key exists but has no expiration set.
	RESP_TTL_NOEXPIRE []byte = []byte(":-1\r\n")
	// RESP_TTL_NOEXIST means the key does not exist or already expired.
	RESP_TTL_NOEXIST []byte = []byte(":-2\r\n")
)

func EvalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'PING' command"), false)
	}

	if len(args) == 0 {
		b = RESP_PONG
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

			// Fixed: evaluate args[i] instead of hardcoded args[3]
			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return resp.Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			exDurationMs = 1000 * exDurationSec
		default:
			return resp.Encode(errors.New("ERR syntax error"), false)
		}
	}
	obj := NewObj(value, exDurationMs)
	Put(key, obj)
	return RESP_OK
}

func EvalGET(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'GET' command"), false)
	}
	key := args[0]
	obj := Get(key)

	if obj == nil {
		return RESP_NIL
	}

	return resp.Encode(obj.Value, false)
}

func EvalTTL(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'TTL' command"), false)
	}
	key := args[0]
	obj := Get(key)

	if obj == nil {
		return RESP_TTL_NOEXIST
	}

	if obj.ExpiresAt == -1 {
		return RESP_TTL_NOEXPIRE
	}

	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		return RESP_TTL_NOEXIST
	}

	return resp.Encode(int64(durationMs/1000), false)
}

func EvalDEL(args []string) []byte {
	if len(args) < 1 {
		return resp.Encode(errors.New("ERR wrong number of arguments for 'DEL' command"), false)
	}

	deleted := Del(args)
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

	result := Expire(key, durationMs)
	return resp.Encode(result, false)
}

func EvalBGREWRITEAOF(args []string) []byte {
	DumpAllAOF()
	return RESP_OK
}

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		log.Println("command : ", cmd.Cmd)
		// Case-insensitive matching
		switch strings.ToUpper(cmd.Cmd) {
		case "PING":
			buf.Write(EvalPING(cmd.Args))
		case "SET":
			buf.Write(EvalSET(cmd.Args))
		case "GET":
			buf.Write(EvalGET(cmd.Args))
		case "TTL":
			buf.Write(EvalTTL(cmd.Args))
		case "DEL":
			buf.Write(EvalDEL(cmd.Args))
		case "EXPIRE":
			buf.Write(EvalEXPIRE(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(EvalBGREWRITEAOF(cmd.Args))
		case "COMMAND":
			buf.Write(RESP_EMPTY_ARRAY)
		default:
			buf.Write(resp.Encode(errors.New("ERR unknown command '"+cmd.Cmd+"'"), false))
		}
	}
	c.Write(buf.Bytes())
}