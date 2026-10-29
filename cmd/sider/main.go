package main

import (
	"context"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Serve()
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Println("Shutdown signal received, closing server...")
		srv.Shutdown()
		if err := <-serverErr; err != nil {
			log.Println("error stopping server:", err)
		}
		log.Println("Taking final RDB snapshot...")
		if err := persistence.TakeSnapshot(); err != nil {
			log.Fatal(err)
		}
	}
}
