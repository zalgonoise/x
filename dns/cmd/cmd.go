package cmd

import (
	"os"
	"strings"

	stdlog "log"

	"github.com/zalgonoise/x/dns/cmd/config"
	svclog "github.com/zalgonoise/x/dns/service/middleware/logger"

	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/jsonfile"
	"github.com/zalgonoise/x/dns/store/memmap"
	"github.com/zalgonoise/x/dns/store/yamlfile"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
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
		apis      endpoints.HTTPAPI
		https     httpapi.Server
	)

	// get config
	conf = ParseFlags()

	// initialize DNS repository
	switch conf.DNS.Type {
	case "miekgdns":
		dnsRepo = core.New(splitFallback(conf.DNS.FallbackDNS)...)
	default:
		dnsRepo = core.New(splitFallback(conf.DNS.FallbackDNS)...)
	}

	// initialize store repository
	switch conf.Store.Type {
	case "memmap":
		storeRepo = memmap.New()
	case "jsonfile":
		storeRepo = jsonfile.New(conf.Store.Path)
	case "yamlfile":
		storeRepo = yamlfile.New(conf.Store.Path)
	default:
		storeRepo = memmap.New()
	}

	// intialize service
	svc = service.New(dnsRepo, storeRepo)

	// setup logger
	var lcfg log.LoggerConfig = nil
	if conf.Logger.Path != "" {
		f, err := os.Open(conf.Logger.Path)
		if err == nil {
			lcfg = log.WithOut(f)
		}
	}
	switch conf.Logger.Type {
	case "text":
		logger = log.New(
			log.CfgTextColorLevelFirst,
			lcfg,
		)
	case "json":
		logger = log.New(
			log.CfgFormatJSON,
			lcfg,
		)
	default:
		logger = nil
	}

	// wrap service in logger, if set
	if logger != nil {
		svc = svclog.LogService(
			svc,
			logger,
		)
	}

	// setup DNS server
	switch conf.DNS.Type {
	case "miekgdns":
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(conf.DNS.Address).
				Prefix(conf.DNS.Prefix).
				Proto(conf.DNS.Proto).
				Build(),
			svc,
		)
	default:
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(conf.DNS.Address).
				Prefix(conf.DNS.Prefix).
				Proto(conf.DNS.Proto).
				Build(),
			svc,
		)
	}

	// setup API endpoints
	apis = endpoints.NewAPI(svc, udps)

	// setup HTTP server
	https = httpapi.NewServer(apis, conf.HTTP.Port)

	// start DNS server if set
	if conf.Autostart.DNS {
		// defer graceful closure
		defer func() {
			err := udps.Stop()
			if err != nil {
				if logger != nil {
					log.Fatalf("error stopping DNS server: %v", err)
				} else {
					stdlog.Fatalf("error stopping DNS server: %v", err)
				}
				os.Exit(1)
			}
		}()

		go func() {
			err := udps.Start()
			if err != nil {
				if logger != nil {
					log.Fatalf("error starting DNS server: %v", err)
				} else {
					stdlog.Fatalf("error starting DNS server: %v", err)
				}
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
