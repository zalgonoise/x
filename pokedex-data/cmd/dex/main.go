package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cli"
	"github.com/zalgonoise/x/pokedex-data/config"
	"github.com/zalgonoise/x/pokedex-data/database"
	"github.com/zalgonoise/x/pokedex-data/pokemon"
)

var modes = []string{"build", "load"}

func main() {
	runner := cli.NewRunner("dex",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"build": cli.Executable(ExecBuild),
			"load":  cli.Executable(ExecLoad),
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

	if _, err := service.Load(ctx, c.Min, c.Max); err != nil {
		return 1, err
	}

	if err := service.Close(); err != nil {
		return 1, err
	}

	return 0, nil
}

func ExecLoad(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("load", flag.ExitOnError)

	uri := fs.String("uri", "", "Postgres DB URI to execute the initial load to")
	input := fs.String("input", "", "path to get the CSV file from. Default is 'pokemon.csv'")
	output := fs.String("output", "", "path to place the CSV file in. Default is 'pokemon.csv'")
	minimum := fs.Int("min", 0, "minimum pokemon ID to query")
	maximum := fs.Int("max", 0, "maximum pokemon ID to query")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	var summaries []pokemon.Summary

	if *input == "" {
		c := cfg.Set[config.Config](
			config.DefaultConfig(),
			config.WithOutput(*output), config.WithMin(*minimum), config.WithMax(*maximum),
		)

		logger.InfoContext(ctx, "setting up CSV w/ local data")
		f, err := os.Create(c.Output)
		if err != nil {
			return 1, err
		}

		service := pokemon.NewService(f)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		logger.InfoContext(ctx, "loading data from source")

		summaries, err = service.Load(ctx, c.Min, c.Max)
		if err != nil {
			return 1, err
		}
	} else {
		f, err := os.Open(*input)
		if err != nil {
			return 1, err
		}

		r := csv.NewReader(f)

		var records = make([][]string, 0, 1025)

		for {
			rec, err := r.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return 1, err
			}

			records = append(records, rec)
		}

		summaries = make([]pokemon.Summary, 0, len(records))

		for i := range records {
			id, err := strconv.Atoi(records[i][0])
			if err != nil {
				return 1, err
			}

			summaries = append(summaries, pokemon.Summary{
				ID:     id,
				Sprite: records[i][1],
				Name:   records[i][2],
			})
		}

		f.Close()
	}

	logger.InfoContext(ctx, "connecting to Postgres")

	db, err := database.OpenPostgres(ctx, *uri, 8, logger)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "running initial load")

	if err = pokemon.Load(ctx, db, summaries); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "operation complete")

	return 0, nil
}
