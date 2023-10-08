package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/steam/cmd/steam/store"
)

const (
	storeOp = "store"
)

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

func run(logger *slog.Logger) (error, int) {
	ctx := context.Background()
	if len(os.Args) <= 1 {
		return errNoOp, 1
	}

	switch os.Args[1] {
	case storeOp:
		return store.Exec(ctx, logger, os.Args[2:])
	default:
		return errInvalidOp, 1
	}
}
