//go:build !linux

package server

import "errors"

// RunTCPAsyncServer is only available on Linux, where the event loop is backed
// by epoll. Other platforms can still use RunTCPSyncServer.
func RunTCPAsyncServer() error {
	return errors.New("asynchronous server (epoll) is only supported on linux")
}
