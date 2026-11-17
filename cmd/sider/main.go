package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for SIDER server")
	flag.IntVar(&config.Port, "port", 7030, "Port number for SIDER server")
	flag.StringVar(&config.RDBFile, "rdb-file", config.RDBFile, "Path of the RDB snapshot file")
	flag.StringVar(&config.CheckpointFile, "checkpoint-file", config.CheckpointFile, "Path of the checkpoint file recording the RDB/AOF boundary")
	flag.Int64Var(&config.SnapshotIntervalMs, "snapshot-interval", config.SnapshotIntervalMs, "Snapshot interval in milliseconds for RDB snapshots")
	flag.Int64Var(&config.AOFMaxSizeBytes, "aof-max-size", config.AOFMaxSizeBytes, "Max size of AOF file in bytes before rewrite")
	flag.Int64Var(&config.AOFRewriteIntervalMs, "aof-rewrite-interval", config.AOFRewriteIntervalMs, "AOF rewrite interval in milliseconds")
	flag.Parse()
}

func serveServer(
	srv *server.Server,
	serverErr chan<- error,
	serverDone chan<- struct{},
) {
	serverErr <- srv.Serve()
	close(serverDone)
}

func startServer(srv *server.Server) (<-chan error, <-chan struct{}) {
	serverErr := make(chan error, 1)
	serverDone := make(chan struct{})

	go serveServer(srv, serverErr, serverDone)

	return serverErr, serverDone
}

func shutdownOnSignal(
	signals <-chan os.Signal,
	srv *server.Server,
	serverDone <-chan struct{},
	shutdownRequested chan<- struct{},
) {
	select {
	case sig := <-signals:
		log.Printf("Shutdown signal received (%s), closing server...", sig)
		srv.Shutdown()
		close(shutdownRequested)
	case <-serverDone:
		// The server stopped before a shutdown signal was received.
	}
}

func waitForShutdown(
	signals <-chan os.Signal,
	srv *server.Server,
	serverDone <-chan struct{},
) <-chan struct{} {
	shutdownRequested := make(chan struct{})

	go shutdownOnSignal(signals, srv, serverDone, shutdownRequested)

	return shutdownRequested
}

func waitForServerShutdown(
	serverErr <-chan error,
	shutdownRequested <-chan struct{},
) (error, bool) {
	select {
	case err := <-serverErr:
		return err, false
	case <-shutdownRequested:
		return <-serverErr, true
	}
}

func main() {
	setupFlags()
	log.Println("Gearing up SIDER...")
	if err := persistence.LoadSnapshot(); err != nil {
		log.Fatal(err)
	}
	if err := persistence.LoadAOF(); err != nil {
		log.Fatal(err)
	}

	srv, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	serverErr, serverDone := startServer(srv)
	shutdownRequestedChan := waitForShutdown(signals, srv, serverDone)

	err, shutdownRequested := waitForServerShutdown(serverErr, shutdownRequestedChan)
	if err != nil {
		log.Fatal(err)
	}

	if shutdownRequested {
		log.Println("Taking final RDB snapshot...")
		if err := persistence.TakeSnapshot(); err != nil {
			log.Fatal(err)
		}
	}
}
