package e2e

import (
	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/health/simplehealth"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
)

var (
	record1 *store.Record = store.New().Type("A").Name("not.a.dom.ain").Addr("192.168.0.10").Build()
	record2 *store.Record = store.New().Type("A").Name("also.not.a.dom.ain").Addr("192.168.0.15").Build()
	record3 *store.Record = store.New().Type("A").Name("really.not.a.dom.ain").Addr("192.168.0.10").Build()
)

func initializeService() service.Service {
	return service.New(
		core.New(),
		memmap.New(),
		simplehealth.New(),
		config.Default(),
	)
}
