package cmd

import (
	"context"
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/secr/cmd/flags"
	"github.com/zalgonoise/x/secr/factory"
)

func Run() {
	// temp logger
	log := logx.Default()

	ctx, s := spanner.Start(context.Background(), "starting secrets server")

	conf := flags.ParseFlags()
	server, err := factory.From(conf)
	if err != nil {
		s.Event("failed to initialize secrets server", attr.String("error", err.Error()))
		s.End()
		log.Fatal("failed to initialize secrets server", attr.String("error", err.Error()))
		os.Exit(1)
	}

	err = server.Start(ctx)
	if err != nil {
		s.Event("failed to start the secrets server", attr.String("error", err.Error()))
		s.End()
		log.Fatal("failed to start the secrets server", attr.String("error", err.Error()))
		os.Exit(1)
	}
	s.End()
}
