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
		core.New(nil), // defaults
		memmap.New(),
	)

	ctx := context.Background()
	err := s.Add(ctx,
		store.New().Type(dns.TypeA.String()).Name("nw.io").Addr("127.0.0.1").Build(),
		store.New().Type(dns.TypeA.String()).Name("host.nw.io").Addr("192.168.0.1").Build(),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
