package main

import (
	"context"
	"fmt"
	"github.com/zalgonoise/x/collide/frontend"
	"github.com/zalgonoise/x/collide/internal/config"
	"go.uber.org/automaxprocs/maxprocs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/collide/internal/log"
)

func main() {
	logger := log.New("debug", true, true)

	runner := cli.NewRunner("collide-fe",
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

func ExecServe(ctx context.Context, logger *slog.Logger, _ []string) (int, error) {
	// init config
	logger.InfoContext(ctx, "loading config")
	cfg, err := config.New()
	if err != nil {
		return 1, err
	}

	_, err = maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		logger.InfoContext(ctx, fmt.Sprintf(s, i...))
	}))
	if err != nil {
		return 1, err
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	serverShutdown := frontend.NewServer(ctx, cfg.Frontend.BackendURI, cfg.Frontend.Port, logger)

	<-signalChannel

	shutdownTimeout := 30 * time.Second

	logger.InfoContext(ctx, "shutting down", slog.Duration("timeout", shutdownTimeout))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := serverShutdown(shutdownCtx); err != nil {
		return 1, err
	}

	return 0, nil
}
