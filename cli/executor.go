package cli

import (
	"context"
	"log/slog"
)

type Executor interface {
	Exec(ctx context.Context, logger *slog.Logger, args []string) (int, error)
}

type Executable func(ctx context.Context, logger *slog.Logger, args []string) (int, error)

func (e Executable) Exec(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	return e(ctx, logger, args)
}
