package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/zalgonoise/x/modupdate/config"
	"github.com/zalgonoise/x/modupdate/events"
)

const (
	actionCheckout   = "checkout"
	actionUpdateRepo = "update.repository"
	actionUpdateMod  = "update.modules"
	actionPushCommit = "push.commit"
	actionPushPush   = "push.push"
)

var ErrBinNotFound = errors.New("binary not found")

type Reporter interface {
	ReportEvent(ctx context.Context, event events.Event)
}

type ModUpdate struct {
	repo     *config.Repository
	checkout *config.Checkout
	update   *config.Update
	push     *config.Push

	reporter Reporter
	logger   *slog.Logger
}

func NewModUpdate(reporter Reporter, cfg *config.Config, logger *slog.Logger) *ModUpdate {
	if cfg == nil {
		return nil
	}

	if reporter == nil {
		reporter = noOpReporter{}
	}

	return &ModUpdate{
		repo:     &cfg.Repository,
		checkout: &cfg.Checkout,
		update:   &cfg.Update,
		push:     &cfg.Push,
		reporter: reporter,
		logger:   logger,
	}
}

func (a *ModUpdate) Run(ctx context.Context) error {
	if err := a.Checkout(ctx); err != nil {
		return err
	}

	if err := a.Update(ctx); err != nil {
		return err
	}

	return a.Push(ctx)
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
