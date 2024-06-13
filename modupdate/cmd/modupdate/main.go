package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"

	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/modupdate/actions"
	"github.com/zalgonoise/x/modupdate/config"
	"github.com/zalgonoise/x/modupdate/database"
	"github.com/zalgonoise/x/modupdate/events"
	"github.com/zalgonoise/x/modupdate/repository"
	"github.com/zalgonoise/x/modupdate/service"
)

var modes = []string{"run"}

func main() {
	runner := cli.NewRunner("modupdate",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"run": cli.Executable(ExecRun),
		}),
	)

	cli.Run(runner)
}

func ExecRun(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	configPath := fs.String("config", "", "path to the config JSON file containing the task definitions as well as database URI and discord token")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	cfg, err := parseConfig(*configPath)
	if err != nil {
		return 1, err
	}

	db, err := database.OpenSQLite(cfg.DatabaseURI, database.ReadWritePragmas(), logger)
	if err != nil {
		return 1, err
	}

	if err = database.MigrateSQLite(ctx, db, logger); err != nil {
		return 1, err
	}

	reporter, err := events.NewReporter(cfg.DiscordToken, logger)
	if err != nil {
		return 1, err
	}

	cron, err := actions.NewActions(reporter, logger, cfg.Tasks...)
	if err != nil {
		return 1, err
	}

	svc, err := service.NewService(cron, repository.NewRepository(db), logger)
	if err != nil {
		return 1, err
	}

	defer func() {
		serviceErr := svc.Err()
		closeErr := svc.Close()

		if err := errors.Join(serviceErr, closeErr); err != nil {
			logger.ErrorContext(ctx, "error when shutting service down", slog.String("error", err.Error()))
		}
	}()

	svc.Start()

	return 0, nil
}

func parseConfig(path string) (*config.Config, error) {
	if path == "" {
		return nil, errors.New("config path cannot be empty")
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{}

	if err := json.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
