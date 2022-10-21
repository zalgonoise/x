package main

import (
	"log"
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/factory"
)

func main() {
	cfg := config.New()

	// create HTTP / UDP server from config
	svr := factory.From(cfg)

	defer func() {
		err := svr.Stop()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}()

	// start HTTP server
	// you need to start DNS server with a GET request to localhost:8080/dns/start
	err := svr.Start()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
