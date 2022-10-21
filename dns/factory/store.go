package factory

import (
	"os"
	"strings"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/health/simplehealth"
	"github.com/zalgonoise/x/dns/service"
	svclog "github.com/zalgonoise/x/dns/service/middleware/logger"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/jsonfile"
	"github.com/zalgonoise/x/dns/store/memmap"
	"github.com/zalgonoise/x/dns/store/yamlfile"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/store/fs"
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

func DNSRepository(rtype string, fallbackDNS ...string) dns.Repository {
	var dnsRepo dns.Repository

	switch rtype {
	case "miekgdns":
		dnsRepo = core.New(fallbackDNS...)
	default:
		dnsRepo = core.New(fallbackDNS...)
	}

	return dnsRepo
}

func HealthRepository(rtype string) health.Repository {
	switch rtype {
	case "simple", "simplehealth":
		return simplehealth.New()
	default:
		return simplehealth.New()
	}
}

func StoreRepository(rtype string, path string) store.Repository {
	var storeRepo store.Repository

	switch rtype {
	case "memmap", "memory", "in-memory":
		storeRepo = memmap.New()
	case "jsonfile", "json":
		storeRepo = jsonfile.New(path)
	case "yamlfile", "yaml":
		storeRepo = yamlfile.New(path)
	default:
		storeRepo = memmap.New()
	}

	return storeRepo
}

func Service(
	dnsRepo dns.Repository,
	storeRepo store.Repository,
	healthRepo health.Repository,
	conf *config.Config,
) service.Service {
	var (
		svc     service.Service = service.New(dnsRepo, storeRepo, healthRepo, conf)
		logConf log.LoggerConfig
		logger  log.Logger
	)

	// short-circuit out
	if conf.Logger.Type == "" {
		return svc
	}

	if conf.Logger.Path != "" {
		_, err := os.Open(conf.Logger.Path)
		switch err {
		case nil:
			f, err := fs.New(conf.Logger.Path)
			if err == nil {
				logConf = log.WithOut(f, os.Stderr)
			}
		default:
			_, err = os.Create(conf.Logger.Path)
			if err == nil {
				f, err := fs.New(conf.Logger.Path)
				if err == nil {
					logConf = log.WithOut(f, os.Stderr)
				}
			}
		}
	}

	switch conf.Logger.Type {
	case "text":
		logger = log.New(
			log.CfgTextColorLevelFirst,
			logConf,
		)
	case "json":
		logger = log.New(
			log.CfgFormatJSON,
			logConf,
		)
	case "csv":
		logger = log.New(
			log.CfgFormatCSV,
			logConf,
		)
	case "xml":
		logger = log.New(
			log.CfgFormatXML,
			logConf,
		)
	default:
		logger = nil
	}

	if logger == nil {
		return svc
	}

	return svclog.LogService(svc, logger)
}

func UDPServer(stype, address, prefix, proto string, svc service.Service) udp.Server {
	var udps udp.Server

	switch stype {
	case "miekgdns":
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(address).
				Prefix(prefix).
				Proto(proto).
				Build(),
			svc,
		)
	default:
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(address).
				Prefix(prefix).
				Proto(proto).
				Build(),
			svc,
		)
	}

	return udps
}

func Server(dnstype, dnsAddress, dnsPrefix, dnsProto string, httpPort int, svc service.Service) (httpapi.Server, udp.Server) {
	udps := UDPServer(dnstype, dnsAddress, dnsPrefix, dnsProto, svc)
	apis := endpoints.NewAPI(svc, udps)
	https := httpapi.NewServer(apis, httpPort)

	return https, udps
}
