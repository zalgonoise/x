package main

import (
	"context"
	"log"

	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
)

func main() {
	// init implementations
	dnscore := core.New()
	memstore := memmap.New()

	// init service
	s := service.New(dnscore, memstore)

	// init transport
	t := miekgdns.New(nil, s)

	ctx := context.Background()
	err := s.AddRecords(ctx,
		store.New().Type(store.TypeA.String()).Name("nw.io").Addr("127.0.0.1").Build(),
		store.New().Type(store.TypeA.String()).Name("host.nw.io").Addr("192.168.0.1").Build(),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Start()
	if err != nil {
		log.Fatal(err)
	}
}
