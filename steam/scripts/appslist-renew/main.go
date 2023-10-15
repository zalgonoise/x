package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/steam"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	logger.InfoContext(ctx, "fetching apps list from Steam")

	if err := steam.GetAppsList(); err != nil {
		logger.ErrorContext(ctx, "failed to fetch apps listing", slog.String("error", err.Error()))

		return
	}

	logger.InfoContext(ctx, "fetched apps list successfully")
}
