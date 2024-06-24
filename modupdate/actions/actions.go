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
	actionCheckout        = "checkout"
	actionCheckoutPresent = "checkout.present"
	actionCheckoutBranch  = "checkout.branch"
	actionUpdateRepo      = "update.repository"
	actionUpdateMod       = "update.modules"
	actionCheckBuild      = "check.build"
	actionPushCommit      = "push.commit"
	actionPushPush        = "push.push"
)

var ErrBinNotFound = errors.New("binary not found")

type Reporter interface {
	ReportEvent(event events.Event)
	Flush()
}

type ModUpdate struct {
	repo     *config.Repository
	checkout *config.Checkout
	update   *config.Update
	check    *config.Check
	push     *config.Push

	reporter Reporter
	logger   *slog.Logger
}

func NewModUpdate(reporter Reporter, cfg *config.Task, logger *slog.Logger) *ModUpdate {
	if cfg == nil {
		return nil
	}

	return &ModUpdate{
		repo:     &cfg.Repository,
		checkout: &cfg.Checkout,
		update:   &cfg.Update,
		check:    &cfg.Check,
		push:     &cfg.Push,
		reporter: reporter,
		logger:   logger,
	}
}

func (a *ModUpdate) Run(ctx context.Context) error {
	defer a.reporter.Flush()

	if err := a.Checkout(ctx); err != nil {
		return err
	}

	if err := a.Update(ctx); err != nil {
		return err
	}

	if err := a.Check(ctx); err != nil {
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
