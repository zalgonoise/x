package factory

import (
	"os"
	"strings"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/zlog/log"
)

func From(conf *config.Config) httpapi.Server {
	// initialize DNS repository
	dnsRepo := DNSRepository(
		conf.DNS.Type,
		strings.Split(conf.DNS.FallbackDNS, ",")...,
	)

	// initialize store repository
	storeRepo := StoreRepository(
		conf.Store.Type,
		conf.Store.Path,
	)

	// initialize health repository
	healthRepo := HealthRepository("")

	// intialize service
	svc := Service(
		dnsRepo,
		storeRepo,
		healthRepo,
		conf,
	)

	// initialize HTTP and DNS servers
	https, udps := Server(
		conf.DNS.Type,
		conf.DNS.Address,
		conf.DNS.Prefix,
		conf.DNS.Proto,
		conf.HTTP.Port,
		svc,
	)

	if conf.Autostart.DNS {
		go func() {
			err := udps.Start()
			if err != nil {
				log.Fatalf("error starting DNS server: %v", err)
				os.Exit(1)
			}
		}()
	}
	return https
}
