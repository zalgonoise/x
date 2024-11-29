package validate

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	obs_midi "github.com/zalgonoise/x/obs-midi"
	"log/slog"
	"os"
)

var ErrSourceEmpty = errors.New("source config cannot be empty")

func Exec(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	source := fs.String("f", "", "path to the target config source file")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *source == "" {
		return 1, ErrSourceEmpty
	}

	data, err := os.ReadFile(*source)
	if err != nil {
		return 1, err
	}

	cfg := &obs_midi.Config{}

	if err := json.Unmarshal(data, cfg); err != nil {
		return 1, err
	}

	if _, err := json.Marshal(obs_midi.NewConfigMap(cfg)); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "config file is valid")

	return 0, nil
}
