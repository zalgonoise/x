package main

import (
	"context"
	"log"

	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
)

func main() {
	memstore := memmap.New()
	dnscore := core.New(memstore)

	s := service.New(
		dnscore,
		memstore,
	)

	ctx := context.Background()
	err := s.Add(ctx, store.Record{
		Name: "nw.io",
		Type: dns.TypeA,
		Addr: "127.0.0.1",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
