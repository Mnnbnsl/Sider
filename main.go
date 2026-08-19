package main

import (
	"flag"
	"log"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for SIDER server")
	flag.IntVar(&config.Port, "port", 7030, "Port number for SIDER server")
	flag.Parse()
}

func main() {	
	setupFlags()
	log.Println("Gearing up SIDER...")
	server.RunTCPSyncServer()
}