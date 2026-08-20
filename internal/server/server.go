package server

import (
	"log"
	"net"
	"strconv"
	"io"

	"github.com/mnnbnsl/sider/config"
)

func readCommand(c net.Conn) (string, error) {

	// reads 512 byte at a time right now
	// TODO : extend it so that larger data is read repeatedly
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func respond(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	} 
	return nil
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
				c.Close()
				con_clients -= 1
				log.Println("Client disconnected :", c.RemoteAddr(), "Concurrent clients :", con_clients)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}
			log.Println("Command", cmd)
			if err = respond(cmd, c); err != nil {
				log.Print("Err write :", err)
			}
		}
	}
}