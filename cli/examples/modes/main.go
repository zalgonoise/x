package main

import (
	"context"
	"flag"
	"log/slog"

	"github.com/zalgonoise/x/cli"
)

func main() {
	runner := cli.NewRunner("printer",
		cli.WithExecutors(map[string]cli.Executor{
			"print":   cli.Executable(ExecPrint),
			"newline": cli.Executable(ExecNewline),
		}),
	)

	cli.Run(runner)
}

func ExecPrint(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("print", flag.ExitOnError)

	input := fs.String("v", "", "the content to print")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *input != "" {
		logger.InfoContext(ctx, "user print request",
			slog.String("data", *input),
		)
	}

	return 0, nil
}

func ExecNewline(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	logger.InfoContext(ctx, "user newline request",
		slog.String("data", "\n\n\n"),
	)

	return 0, nil
}
