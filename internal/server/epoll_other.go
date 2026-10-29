//go:build !linux

package server

import "errors"

type Server struct{}

func New() (*Server, error) {
	return nil, errors.New("asynchronous server (epoll) is only supported on linux")
}

func (s *Server) Serve() error {
	return errors.New("asynchronous server (epoll) is only supported on linux")
}

func (s *Server) Shutdown() {}
