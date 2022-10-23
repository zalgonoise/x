package cmd

import (
	"os"

	"log"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/cmd/flags"
	"github.com/zalgonoise/x/dns/factory"

	"github.com/zalgonoise/x/dns/transport/httpapi"
)

// Run starts the DNS app based on the input configuration (file configuration,
// CLI flags, and OS environment variables)
//
// Blocking call; will error out (with os.Exit(1)) if failed
func Run() {
	var (
		conf *config.Config
		svr  httpapi.Server
	)

	// get config
	conf = flags.ParseFlags()

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
