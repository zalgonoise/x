package main

import (
	"os"

	"github.com/zalgonoise/zlog/log"

	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/service/middleware/logger"
	"github.com/zalgonoise/x/dns/store/jsonfile"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
)

func main() {
	// init implementations
	dnsCore := core.New("1.1.1.1") // falls back to one-dot
	jsonMemStore := jsonfile.New("/tmp/dns/dns.json")

	// init service
	s := service.New(dnsCore, jsonMemStore)
	loggedSvc := logger.LogService(
		s,
		log.New(
			log.CfgTextColorLevelFirst,
		),
	)

	// init UDP server
	udps := miekgdns.NewServer(
		udp.NewDNS().Build(),
		loggedSvc,
	)

	// init API endpoints
	apis := endpoints.NewAPI(loggedSvc, udps)

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
