package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/core"
	"github.com/mnnbnsl/sider/internal/resp"
)

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {

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

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.RedisCmd, c io.ReadWriter) {
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



func RunTCPAsyncServer() error {
	log.Println("starting an asynchronous TCP server on", config.Host, config.Port)
	con_clients := 0

	max_clients := 20000
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	// Create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// creating EPOLL instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// Specify the events we want to get hints about
	// and set the socket on which
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Listen to read events on the Server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	for {
		// see if any FD is ready for an IO
		nevents, err := syscall.EpollWait(epollFD, events[:], 100)
		if err != nil {
			continue
		}
		if nevents == 0 {
			core.Cleanup()
			continue
		}

		for i := 0; i < nevents; i++ {
			// if the socket server itself is ready for an IO
			if int(events[i].Fd) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				// increase the number of concurrent clients count
				con_clients++
				syscall.SetNonblock(fd, true)

				// add this new TCP connection to be monitored
				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}
				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					con_clients -= 1
					continue
				}
				respond(cmd, comm)
			}
		}
	}
}