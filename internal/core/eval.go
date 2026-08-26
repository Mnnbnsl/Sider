package core 

import (
	"log"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/mnnbnsl/sider/internal/resp"	
)

var RESP_NIL []byte = []byte("$-1\r\n")

func EvalPING(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = resp.Encode("PONG", true)
	} else {
		b = resp.Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func EvalSET(args []string, c io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}
	var exDurationMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i]{
		case "ex", "EX":
			i++
			if i == len(args) {
				return errors.New("(error) ERR syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return errors.New("(error) ERR value is not an integer or out of range")
			}
			exDurationMs = 1000 * exDurationSec
		default:
			return errors.New("(error) ERR syntax error")
		}
	}
	obj := NewObj(value, exDurationMs)
	Put(key, obj)
	c.Write([]byte("+OK\r\n"))
	return nil
}

func EvalGET(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'get' command")
	}
	key := args[0]
	obj := Get(key)

	// object was nil
	if obj == nil {
		c.Write(RESP_NIL)
		return nil
	}

	// object expired
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli(){
		c.Write(RESP_NIL)
		return nil 
	}

	c.Write(resp.Encode(obj.Value, false))
	return nil 
}

func EvalTTL(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'ttl' command")
	}
	key := args[0]
	obj := Get(key)

	// object was nil
	if obj == nil {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	// No expiration
	if obj.ExpiresAt == -1 {
		c.Write([]byte(":-1\r\n"))
		return nil
	}

	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	c.Write(resp.Encode(int64(durationMs/1000), false))
	return nil 
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {
	log.Println("command : ", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return EvalPING(cmd.Args, c)
	case "SET":
		return EvalSET(cmd.Args, c)
	case "GET" :
		return EvalGET(cmd.Args, c)
	case "TTL" :
		return EvalTTL(cmd.Args, c)
	default:
		return errors.New("ERR unknown command")
	}
}