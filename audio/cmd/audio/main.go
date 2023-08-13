package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/audio/cmd/audio/config"
	"github.com/zalgonoise/x/audio/cmd/audio/stream"
)

func main() {
	err, code := run()
	if err != nil {
		slog.Error(
			"audio/gsp: runtime error",
			slog.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (error, int) {
	cfg, err := config.WithDefaults()
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
