package monitor

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/steam/cmd/steam/alert"

	"github.com/zalgonoise/x/cron"
	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/cron/schedule"
	"github.com/zalgonoise/x/cron/selector"
)

const defaultSchedule = "0 10 * * *"

var (
	errEmptyID          = errors.New("empty app ID")
	errEmptyURL         = errors.New("empty webhook URL")
	errEmptyTargetPrice = errors.New("empty target price")
)

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("alert", flag.ExitOnError)

	ids := fs.String("ids", "", "comma-separated list of app ID values")
	country := fs.String("country", "", "country code (2-character-long)")
	platform := fs.String("platform", "logger", "target platform where to post (logger; slack; discord)")
	url := fs.String("url", "", "webhook target URL (platform: slack; discord)")
	targetDiscount := fs.Int("target_discount", 50, "target discount percent")
	cronSchedule := fs.String("schedule", defaultSchedule, fmt.Sprintf("schedule frequency to query for discounts, as a cron schedule string (default: %s)", defaultSchedule))

	if err := fs.Parse(args); err != nil {
		return err, 1
	}

	if err := fs.Parse(args); err != nil {
		return err, 1
	}

	if *ids == "" {
		return errEmptyID, 1
	}

	if *targetDiscount == 0 {
		return errEmptyTargetPrice, 1
	}

	if *platform == "" {
		*platform = "logger"
	}

	if *url == "" && *platform != "logger" {
		return errEmptyURL, 1
	}

	if *cronSchedule == "" {
		*cronSchedule = defaultSchedule
	}

	return runCron(ctx, logger, *cronSchedule, *ids, *country, *platform, *url, *targetDiscount)
}

func runCron(
	ctx context.Context, logger *slog.Logger,
	cronSchedule, ids, country, platform, url string,
	targetDiscount int,
) (error, int) {
	runner := func(execCtx context.Context) error {
		return alert.QueryPrices(execCtx, logger, ids, country, platform, url, targetDiscount)
	}

	s, err := schedule.New(schedule.WithSchedule(cronSchedule))
	if err != nil {
		return err, 1
	}

	e, err := executor.New(
		"steam_discount_monitor",
		executor.WithScheduler(s),
		executor.WithRunners(executor.Runnable(runner)),
		executor.WithLogs(logger),
	)
	if err != nil {
		return err, 1
	}

	sel, err := selector.New(
		selector.WithExecutors(e),
		selector.WithLogs(logger),
	)
	if err != nil {
		return err, 1
	}

	c, err := cron.New(cron.WithSelector(sel), cron.WithLogs(logger))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close(signalCh)

	errs := c.Err()

	go c.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil, 0
		case err = <-errs:
			return err, 1
		case sig := <-signalCh:
			return fmt.Errorf("process stopped with OS signal: %s", sig.String()), 0
		}
	}
}
