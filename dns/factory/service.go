package factory

import (
	"os"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/service"
	svclog "github.com/zalgonoise/x/dns/service/middleware/logger"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/store/fs"
)

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
