package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/command"
	"github.com/mnnbnsl/sider/internal/resp"
)

func readCommands(c io.ReadWriter) (command.RedisCmds, error) {
	// reads 512 byte at a time right now
	// TODO : extend it so that larger data is read repeatedly
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}

	values, err := resp.Decode(buf[:n])
	if err != nil {
		return nil, err
	}

	var cmds []*command.RedisCmd = make([]*command.RedisCmd, 0)
	for _, value := range values {
		tokens, err := resp.ToArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &command.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}
	return cmds, nil
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmds command.RedisCmds, c io.ReadWriter) {
	command.EvalAndRespond(cmds, c)
}

func RunTCPSyncServer() {
	log.Println("Starting a Synchronous TCP server on", config.Host, config.Port)

	var con_clients = 0

	lsnr, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	// infinite loop
	for {

		c, err := lsnr.Accept()
		if err != nil {
			panic(err)
		}

		con_clients += 1
		log.Println("Client connected with address :", c.RemoteAddr(), "Concurrent clients :", con_clients)

		// another infinite loop
		for {
			// over the socket, read the command and print it as it is for now
			cmds, err := readCommands(c)
			if err != nil {
				if err == io.EOF {
					c.Close()
					con_clients -= 1
					log.Println("Client disconnected :", c.RemoteAddr(), "Concurrent clients :", con_clients)
					break
				}
				log.Println("err", err)
				respondError(err, c)
				continue
			}
			respond(cmds, c)
		}
	}
}
