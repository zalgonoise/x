package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/steam/apps"
)

const topLevelURI = "internal/app_list/applist.db"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	logger.InfoContext(ctx, "rebuilding local apps list database")

	if _, err := apps.NewIndexer(topLevelURI, logger); err != nil {
		logger.ErrorContext(ctx, "failed to setup database", slog.String("error", err.Error()))

		return
	}

	logger.InfoContext(ctx, "database setup successfully")
}
