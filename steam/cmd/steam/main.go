package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/steam/cmd/steam/alert"
	"github.com/zalgonoise/x/steam/cmd/steam/monitor"
	"github.com/zalgonoise/x/steam/cmd/steam/search"
	"github.com/zalgonoise/x/steam/cmd/steam/store"
)

const (
	storeOp   = "store"
	alertOp   = "alert"
	monitorOp = "monitor"
	searchOp  = "search"
)

var allOperations = []string{
	storeOp, alertOp, monitorOp, searchOp,
}

var (
	errNoOp      = errors.New("undefined operation")
	errInvalidOp = errors.New("invalid operation")
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	err, code := run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
		os.Exit(code)
	}

	os.Exit(code)
}

func printHelp(ctx context.Context, logger *slog.Logger, err error) (error, int) {
	logger.InfoContext(ctx, "please use one of the supported operations",
		slog.Any("operations", allOperations),
		slog.String("error", err.Error()),
	)

	return nil, 1
}

func run(logger *slog.Logger) (error, int) {
	ctx := context.Background()
	if len(os.Args) <= 1 {
		return printHelp(ctx, logger, errNoOp)
	}

	switch os.Args[1] {
	case storeOp:
		return store.Exec(ctx, logger, os.Args[2:])
	case alertOp:
		return alert.Exec(ctx, logger, os.Args[2:])
	case monitorOp:
		return monitor.Exec(ctx, logger, os.Args[2:])
	case searchOp:
		return search.Exec(ctx, logger, os.Args[2:])
	default:
		return printHelp(ctx, logger, errInvalidOp)
	}
}
