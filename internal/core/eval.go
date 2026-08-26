package core 

import (
	"log"
	"errors"
	"net"

	"github.com/mnnbnsl/sider/internal/resp"	
)

func EvalPING(args []string, c net.Conn) error {
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

func EvalAndRespond(cmd *RedisCmd, c net.Conn) error {
	log.Println("command : ", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return EvalPING(cmd.Args, c)
	default:
		return errors.New("ERR unknown command")
	}
}