package main

import (
	"context"
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"

	"github.com/zalgonoise/x/audio/cmd/audio/config"
	"github.com/zalgonoise/x/audio/cmd/audio/stream"
)

func main() {
	err, code := run()
	if err != nil {
		logx.New(texth.New(os.Stderr)).Error(
			"audio/gsp: runtime error",
			attr.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (error, int) {
	cfg, err := config.NewConfig()
	if err != nil {
		return err, 1
	}

	s, err := stream.New(cfg)
	if err != nil {
		return err, 1
	}

	ctx := context.Background()

	err = s.Run(ctx)
	if err != nil {
		return err, 1
	}

	err = s.Close()
	if err != nil {
		return err, 1
	}

	return nil, 0
}
