package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/pokedex-data/config"
	"github.com/zalgonoise/x/pokedex-data/pokemon"
)

var modes = []string{"build"}

func main() {
	runner := cli.NewRunner("dex",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"build": cli.Executable(ExecBuild),
		}),
	)

	cli.Run(runner)
}

func ExecBuild(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	output := fs.String("output", "", "path to place the CSV file in. Default is 'pokemon.csv'")
	minimum := fs.Int("min", 0, "minimum pokemon ID to query")
	maximum := fs.Int("max", 0, "maximum pokemon ID to query")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	c := cfg.Set[config.Config](
		config.DefaultConfig(),
		config.WithOutput(*output), config.WithMin(*minimum), config.WithMax(*maximum),
	)

	f, err := os.Create(c.Output)
	if err != nil {
		return 1, err
	}

	service := pokemon.NewService(f)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := service.Load(ctx, c.Min, c.Max); err != nil {
		return 1, err
	}

	if err := service.Close(); err != nil {
		return 1, err
	}

	return 0, nil
}
