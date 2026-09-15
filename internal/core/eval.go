package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/mnnbnsl/sider/internal/resp"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func EvalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'PING' command"), false)
	}

	if len(args) == 0 {
		b = resp.Encode("PONG", true)
	} else {
		b = resp.Encode(args[0], false)
	}

	return b
}

func EvalSET(args []string) []byte {
	if len(args) <= 1 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'SET' command"), false)
	}
	var exDurationMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i]{
		case "ex", "EX":
			i++
			if i == len(args) {
				return resp.Encode(errors.New("(error) syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return resp.Encode(errors.New("(error) value is not an integer or out of range"),false)
			}
			exDurationMs = 1000 * exDurationSec
		default:
			return resp.Encode(errors.New("(error) syntax error"), false)
		}
	}
	obj := NewObj(value, exDurationMs)
	Put(key, obj)
	return []byte("+OK\r\n")
	
}

func EvalGET(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'GET' command"),false)
	}
	key := args[0]
	obj := Get(key)

	// object was nil or expired
	if obj == nil {
		return RESP_NIL
	}

	return resp.Encode(obj.Value, false)
}

func EvalTTL(args []string) []byte {
	if len(args) != 1 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'TTL' command"), false)
	}
	key := args[0]
	obj := Get(key)

	// object not present
	if obj == nil {
		return []byte(":-2\r\n")
	}

	// No expiration
	if obj.ExpiresAt == -1 {
		return []byte(":-1\r\n")
	}

	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		return []byte(":-2\r\n")
	}

	return resp.Encode(int64(durationMs/1000), false)
}

func EvalDEL(args []string) []byte {
	if len(args) < 1 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'DEL' command"), false)
	}
	
	deleted := Del(args)
	return resp.Encode(deleted, false)
}

func EvalEXPIRE(args []string) []byte {
	if len(args) != 2 {
		return resp.Encode(errors.New("(error) wrong number of arguments for 'EXPIRE' command"), false)
	}

	key := args[0]
	durationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return resp.Encode(errors.New("(error) value is not an integer or out of range"), false)
	}
	
	durationMs := durationSec * 1000

	result := Expire(key, durationMs)
	return resp.Encode(result, false)
}

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		log.Println("command : ", cmd.Cmd)
		switch cmd.Cmd {
		case "PING":
			buf.Write(EvalPING(cmd.Args))
		case "SET":
			buf.Write(EvalSET(cmd.Args))
		case "GET" :
			buf.Write(EvalGET(cmd.Args))
		case "TTL" :
			buf.Write(EvalTTL(cmd.Args))
		case "DEL" :
			buf.Write(EvalDEL(cmd.Args))
		case "EXPIRE" :
			buf.Write(EvalEXPIRE(cmd.Args))
		default:
			buf.Write(resp.Encode(errors.New("(error) unknown command"), false))
		}
	}
	c.Write(buf.Bytes())
}