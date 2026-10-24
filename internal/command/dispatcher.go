package command

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/mnnbnsl/sider/internal/resp"
)

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
			buf.Write(resp.RESP_EMPTY_ARRAY)
		default:
			buf.Write(resp.Encode(errors.New("ERR unknown command '"+cmd.Cmd+"'"), false))
		}
	}
	c.Write(buf.Bytes())
}
