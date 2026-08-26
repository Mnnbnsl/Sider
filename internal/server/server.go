package server

import (
	"log"
	"net"
	"strconv"
	"io"
	"fmt"
	"strings"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/core"
	"github.com/mnnbnsl/sider/internal/resp"
)

func readCommand(c net.Conn) (*core.RedisCmd, error) {

	// reads 512 byte at a time right now
	// TODO : extend it so that larger data is read repeatedly
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}
	
	tokens, err := resp.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err 
	}

	return &core.RedisCmd{
		Cmd : strings.ToUpper(tokens[0]),
		Args : tokens[1:],
	}, nil
}

func respondError(err error, c net.Conn) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.RedisCmd, c net.Conn) {
	if err := core.EvalAndRespond(cmd, c); err != nil {
		respondError(err, c)
	} 
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
			cmd, err := readCommand(c)
			if err != nil {
				if err == io.EOF {
					c.Close()
					con_clients -= 1
					log.Println("Client disconnected :", c.RemoteAddr(), "Concurrent clients :", con_clients,)
					break
				}
				log.Println("err", err)
				respondError(err, c)
				continue
			}
			respond(cmd, c)
		}
	}
}