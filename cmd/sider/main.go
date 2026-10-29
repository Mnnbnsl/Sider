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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		log.Println("Shutdown signal received, taking RDB snapshot...")
		if err := persistence.TakeSnapshot(); err != nil {
			log.Println("error taking RDB snapshot on shutdown:", err)
		}
		os.Exit(0)
	}()

	if err := server.RunTCPAsyncServer(); err != nil {
		log.Fatal(err)
	}
}
