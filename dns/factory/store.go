package factory

import (
	"os"

	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/dns/core"
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

func StoreRepository(rtype string, path string) store.Repository {
	var storeRepo store.Repository

	switch rtype {
	case "memmap":
		storeRepo = memmap.New()
	case "jsonfile":
		storeRepo = jsonfile.New(path)
	case "yamlfile":
		storeRepo = yamlfile.New(path)
	default:
		storeRepo = memmap.New()
	}

	return storeRepo
}

func Service(dnsRepo dns.Repository, storeRepo store.Repository, ltype string, path string) service.Service {
	var (
		svc     service.Service = service.New(dnsRepo, storeRepo)
		logConf log.LoggerConfig
		logger  log.Logger
	)

	// short-circuit out
	if ltype == "" {
		return svc
	}

	if path != "" {
		_, err := os.Open(path)
		if err == nil {
			if f, err := fs.New(path); err == nil {
				logConf = log.WithOut(f)
			}
		}
	}

	switch ltype {
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
