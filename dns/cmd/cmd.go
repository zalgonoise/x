package cmd

import (
	"os"
	"strings"

	stdlog "log"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/factory"

	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/zlog/log"
)

func Run() {
	var (
		conf *config.Config

		dnsRepo   dns.Repository
		storeRepo store.Repository
		svc       service.Service
		logger    log.Logger
		udps      udp.Server
		https     httpapi.Server
	)

	// get config
	conf = ParseFlags()

	// initialize DNS repository
	dnsRepo = factory.DNSRepository(conf.DNS.Type, splitFallback(conf.DNS.FallbackDNS)...)

	// initialize store repository
	storeRepo = factory.StoreRepository(conf.Store.Type, conf.Store.Path)

	// intialize service
	svc = factory.Service(dnsRepo, storeRepo, conf.Logger.Type, conf.Logger.Path)

	// setup HTTP and DNS servers
	https, udps = factory.Server(
		conf.DNS.Type,
		conf.DNS.Address,
		conf.DNS.Prefix,
		conf.DNS.Proto,
		conf.HTTP.Port,
		svc,
	)

	if conf.Autostart.DNS {
		// defer graceful closure
		defer func() {
			err := udps.Stop()
			if err != nil {
				log.Fatalf("error stopping DNS server: %v", err)
				os.Exit(1)
			}
		}()

		go func() {
			err := udps.Start()
			if err != nil {
				log.Fatalf("error starting DNS server: %v", err)
				os.Exit(1)
			}
		}()
	}

	// start HTTP server
	// defer graceful closure
	defer func() {
		err := https.Stop()
		if err != nil {
			if logger != nil {
				log.Fatalf("error stopping HTTP server: %v", err)
			} else {
				stdlog.Fatalf("error stopping HTTP server: %v", err)
			}
			os.Exit(1)
		}
	}()

	err := https.Start()
	if err != nil {
		if logger != nil {
			log.Fatalf("error starting HTTP server: %v", err)
		} else {
			stdlog.Fatalf("error starting HTTP server: %v", err)
		}
		os.Exit(1)
	}
}

func splitFallback(s string) []string {
	if len(s) > 14 {
		return strings.Split(s, ",")
	}
	return []string{s}
}
