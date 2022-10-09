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
	s := service.New(
		core.New(),
		memmap.New(),
	)

	ctx := context.Background()
	err := s.Add(ctx,
		store.Record{
			Name: "nw.io",
			Type: dns.TypeA.String(),
			Addr: "127.0.0.1",
		},
		store.Record{
			Name: "host.nw.io",
			Type: dns.TypeA.String(),
			Addr: "192.168.0.1",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
