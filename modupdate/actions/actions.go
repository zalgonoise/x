package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/zalgonoise/x/modupdate/config"
)

var ErrBinNotFound = errors.New("binary not found")

type ModUpdate struct {
	repo     *config.Repository
	checkout *config.Checkout
	update   *config.Update
	push     *config.Push

	logger *slog.Logger
}

func NewModUpdate(cfg *config.Config, logger *slog.Logger) *ModUpdate {
	if cfg == nil {
		return nil
	}

	return &ModUpdate{
		repo:     &cfg.Repository,
		checkout: &cfg.Checkout,
		update:   &cfg.Update,
		push:     &cfg.Push,
		logger:   logger,
	}
}

func cmd(ctx context.Context, dir, bin string, args ...string) ([]string, error) {
	binPath := bin

	if binPath != "" && binPath[0] != '/' {
		binPath = getPath(ctx, bin)

		if binPath == "" {
			return nil, fmt.Errorf("%w: %s", ErrBinNotFound, bin)
		}
	}

	c := exec.CommandContext(ctx, binPath, args...)

	if dir != "" {
		c.Dir = dir
	}

	buf, err := c.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		return []string{}, nil
	}

	split := strings.Split(string(buf), "\n")

	output := make([]string, 0, len(split))

	for i := range split {
		if split[i] != "" {
			output = append(output, split[i])
		}
	}

	return output, nil
}
