//go:build linux

package server

import (
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/store"
)

type Server struct {
	listenerFD  int
	epollFD     int
	wakeReadFD  int
	wakeWriteFD int

	mu      sync.Mutex
	clients map[int]struct{}
	stop    sync.Once
}

func New() (*Server, error) {
	log.Println("starting an asynchronous TCP server on", config.Host, config.Port)

	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}
	closeServerFD := func() {
		syscall.Close(serverFD)
	}

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		closeServerFD()
		return nil, err
	}

	ip4 := net.ParseIP(config.Host)
	if ip4 == nil || ip4.To4() == nil {
		closeServerFD()
		return nil, syscall.EINVAL
	}
	var addr [4]byte
	copy(addr[:], ip4.To4())
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: addr,
	}); err != nil {
		closeServerFD()
		return nil, err
	}
	if err = syscall.Listen(serverFD, 20000); err != nil {
		closeServerFD()
		return nil, err
	}

	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		closeServerFD()
		return nil, err
	}
	closeResources := func() {
		syscall.Close(epollFD)
		closeServerFD()
	}

	wakeFDs := make([]int, 2)
	if err = syscall.Pipe(wakeFDs); err != nil {
		closeResources()
		return nil, err
	}
	closeAll := func() {
		syscall.Close(wakeFDs[0])
		syscall.Close(wakeFDs[1])
		closeResources()
	}

	s := &Server{
		listenerFD:  serverFD,
		epollFD:     epollFD,
		wakeReadFD:  wakeFDs[0],
		wakeWriteFD: wakeFDs[1],
		clients:     make(map[int]struct{}),
	}

	addEvent := func(fd int) error {
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(fd),
		})
	}
	if err = addEvent(serverFD); err != nil {
		closeAll()
		return nil, err
	}
	if err = addEvent(wakeFDs[0]); err != nil {
		closeAll()
		return nil, err
	}
	return s, nil
}

func (s *Server) Serve() error {
	defer func() {
		syscall.Close(s.listenerFD)
		syscall.Close(s.epollFD)
		syscall.Close(s.wakeReadFD)
		syscall.Close(s.wakeWriteFD)
		s.mu.Lock()
		for fd := range s.clients {
			syscall.Close(fd)
			delete(s.clients, fd)
		}
		s.mu.Unlock()
	}()

	events := make([]syscall.EpollEvent, 20000)
	for {
		nevents, err := syscall.EpollWait(s.epollFD, events, 100)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		if nevents == 0 {
			store.Cleanup()
			persistence.MaybeSnapshot()
			persistence.MaybeRewriteAOF()
			continue
		}

		for i := 0; i < nevents; i++ {
			fd := int(events[i].Fd)
			if fd == s.wakeReadFD {
				return nil
			}
			if fd == s.listenerFD {
				clientFD, _, err := syscall.Accept(s.listenerFD)
				if err != nil {
					if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
						continue
					}
					return err
				}
				if err := syscall.SetNonblock(clientFD, true); err != nil {
					syscall.Close(clientFD)
					continue
				}
				if err := syscall.EpollCtl(s.epollFD, syscall.EPOLL_CTL_ADD, clientFD, &syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(clientFD),
				}); err != nil {
					syscall.Close(clientFD)
					continue
				}
				s.mu.Lock()
				s.clients[clientFD] = struct{}{}
				s.mu.Unlock()
				continue
			}

			comm := FDComm{Fd: fd}
			cmds, err := readCommands(comm)
			if err != nil {
				s.closeClient(fd)
				continue
			}
			respond(cmds, comm)
		}
	}
}

func (s *Server) closeClient(fd int) {
	syscall.EpollCtl(s.epollFD, syscall.EPOLL_CTL_DEL, fd, nil)
	syscall.Close(fd)
	s.mu.Lock()
	delete(s.clients, fd)
	s.mu.Unlock()
}

func (s *Server) Shutdown() {
	s.stop.Do(func() {
		syscall.Close(s.listenerFD)
		s.mu.Lock()
		for fd := range s.clients {
			syscall.Close(fd)
		}
		s.mu.Unlock()
		_, _ = syscall.Write(s.wakeWriteFD, []byte{1})
	})
}
