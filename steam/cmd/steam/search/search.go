package search

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/zalgonoise/x/steam/apps"
)

var (
	errEmptyName = errors.New("name cannot be empty")
)

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)

	name := fs.String("name", "", "name of the game or product")
	dbURI := fs.String("db", "", "path to read and write app list data, in a SQLite DB file")

	if err := fs.Parse(args); err != nil {
		return err, 1
	}

	if *name == "" {
		return errEmptyName, 1
	}

	r, err := apps.NewRepository(apps.WithURI(*dbURI))
	if err != nil {
		return err, 1
	}

	results, err := r.Search(ctx, *name)
	if err != nil {
		return err, 1
	}

	for i := range results {
		logger.InfoContext(ctx, fmt.Sprintf("result #%d", i), slog.String("name", results[i].Name), slog.Int("app_id", int(results[i].AppID)))
	}

	return nil, 0
}
