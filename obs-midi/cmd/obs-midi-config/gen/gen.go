package gen

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	obs_midi "github.com/zalgonoise/x/obs-midi"
	"log/slog"
	"os"
	"slices"
	"strings"
)

var (
	ErrSourceEmpty      = errors.New("source config cannot be empty")
	ErrColorUnsupported = errors.New("color schema is not supported")
)

var supportedColors = []string{"", "default", "green"}

func Exec(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	source := fs.String("f", "", "path to the target config source file")
	out := fs.String("o", "obs-midi-config-gen.json", "path to the destination config file")
	color := fs.String("c", "green", "color schema enumeration (default, green)")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	if *source == "" {
		return 1, ErrSourceEmpty
	}

	if !slices.Contains(supportedColors, strings.ToLower(*color)) {
		return 1, ErrColorUnsupported
	}

	data, err := os.ReadFile(*source)
	if err != nil {
		return 1, err
	}

	cfg := &obs_midi.Config{}

	if err := json.Unmarshal(data, cfg); err != nil {
		return 1, err
	}

	switch *color {
	case "":
	case "default":
		logger.InfoContext(ctx, "overriding color schema with default colors",
			slog.Any("original_color_schema", cfg.ColorSchema),
		)

		cfg.ColorSchema = obs_midi.DefaultColorSchema()
	case "green":
		logger.InfoContext(ctx, "overriding color schema with green color schema",
			slog.Any("original_color_schema", cfg.ColorSchema),
		)

		cfg.ColorSchema = obs_midi.GreenColorSchema()
	}

	output, err := json.Marshal(obs_midi.NewConfigMap(cfg))
	if err != nil {
		return 1, err
	}

	f, err := os.Create(*out)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "output file created",
		slog.String("path", *out),
	)

	defer f.Close()

	if _, err := f.Write(output); err != nil {
		return 1, err
	}

	return 0, nil
}
