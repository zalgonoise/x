package cmd

import (
	"os"
	"strings"

	"log"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/factory"

	"github.com/zalgonoise/x/dns/transport/httpapi"
)

func Run() {
	var (
		conf *config.Config
		svr  httpapi.Server
	)

	// get config
	conf = ParseFlags()

	// create HTTP / UDP servers
	svr = factory.From(conf)

	// start HTTP server
	// defer graceful closure
	defer func() {
		err := svr.Stop()
		if err != nil {
			log.Fatalf("error stopping HTTP server: %v", err)
			os.Exit(1)
		}
	}()

	err := svr.Start()
	if err != nil {
		log.Fatalf("error starting HTTP server: %v", err)
		os.Exit(1)
	}
}

func splitFallback(s string) []string {
	if len(s) > 14 {
		return strings.Split(s, ",")
	}
	return []string{s}
}
