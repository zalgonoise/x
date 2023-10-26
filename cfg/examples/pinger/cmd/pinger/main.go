package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/cfg/examples/pinger/ping"
)

func main() {
	err, code := run()

	if err != nil {
		slog.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func run() (error, int) {
	myURL := "https://github.com/"
	ctx := context.Background()

	c, err := ping.NewChecker(
		ping.WithURL(myURL),
		// in this case the service has a default for the timeout, but we could
		// override that value if WithTimeout below was not commented out.
		//
		// ping.WithTimeout(30 * time.Second),
	)
	if err != nil {
		return err, 1
	}

	ok, err := c.Up(ctx)
	if err != nil {
		return err, 1
	}

	switch ok {
	case true:
		slog.InfoContext(ctx, "service is up")
	default:
		slog.WarnContext(ctx, "service isn't up")
	}

	return nil, 0
}
