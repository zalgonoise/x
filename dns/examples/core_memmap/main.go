package main

import (
	"log"
	"os"

	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store/memmap"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
)

func main() {
	// init implementations
	dnscore := core.New()
	memstore := memmap.New()

	// init service
	s := service.New(dnscore, memstore)

	// init UDP server
	udps := miekgdns.NewServer(
		udp.NewDNS().Build(),
		s,
	)

	// init API endpoints
	apis := endpoints.NewAPI(s, udps)

	// init HTTP server (defer graceful closure)
	svr := httpapi.NewServer(apis, 8080)
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
