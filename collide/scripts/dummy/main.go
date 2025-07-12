package main

import (
	"context"
	"flag"
	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/collide/internal/log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var modes = []string{"load"}

func main() {
	logger := log.New("debug", true, true)

	runner := cli.NewRunner("dummy",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"load": cli.Executable(ExecLoad),
		}),
	)

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func ExecLoad(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("n", flag.ExitOnError)

	n := fs.Int("n", 20, "number of HTTP requests to send to each endpoint")
	dur := fs.Duration("timeout", 30*time.Second, "maximum duration for the load to run before timing out")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	endpoints := []string{
		"http://localhost:8083/v1/collide/districts",
		"http://localhost:8083/v1/collide/districts/Waterfront/all",
		"http://localhost:8083/v1/collide/districts/Waterfront/drift",
		"http://localhost:8083/v1/collide/districts/Waterfront/all/Container/alternatives",
		"http://localhost:8083/v1/collide/districts/Waterfront/all/Container/collisions",
	}

	ctx, cancel := context.WithTimeout(ctx, *dur)
	defer cancel()

	client := &http.Client{Timeout: *dur}

	for range *n {
		for _, endpoint := range endpoints {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
			if err != nil {
				logger.WarnContext(ctx, "preparing request",
					slog.String("error", err.Error()),
					slog.String("endpoint", endpoint))

				continue
			}

			res, err := client.Do(req)
			if err != nil {
				logger.WarnContext(ctx, "sending request to API",
					slog.String("error", err.Error()),
					slog.String("endpoint", endpoint))

				continue
			}

			if err := res.Body.Close(); err != nil {
				logger.WarnContext(ctx, "closing request body",
					slog.String("error", err.Error()),
					slog.String("endpoint", endpoint))

				continue
			}

			if res.StatusCode > 399 {
				logger.WarnContext(ctx, "analyzing HTTP response",
					slog.Int("status_code", res.StatusCode),
					slog.String("endpoint", endpoint))

				continue
			}

			logger.InfoContext(ctx, "API call completed", slog.String("endpoint", endpoint))
		}
	}

	logger.InfoContext(ctx, "sent requests to all endpoints successfully")

	return 0, nil
}
