package main

import (
	"context"
	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/collide/internal/log"
	"log/slog"
	"os"
)

var modes = []string{"ca", "authz"}

func main() {
	logger := log.New("debug")

	runner := cli.NewRunner("collide",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"serve": cli.Executable(ExecServe),
		}),
	)

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func ExecServe(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	return 0, nil
}
