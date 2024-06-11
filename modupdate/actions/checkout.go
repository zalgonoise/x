package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zalgonoise/x/modupdate/config"
)

var ErrEmptyPath = errors.New("target path to checkout into is empty")

func (a *ModUpdate) Checkout(ctx context.Context) error {
	if a.checkout.Path == "" {
		if a.checkout.Persist {
			a.logger.WarnContext(ctx, "cannot persist with an empty target path",
				slog.String("error", ErrEmptyPath.Error()))

			a.checkout.Persist = false
		}

		path, err := os.MkdirTemp("/tmp", "modupdate-*")
		if err != nil {
			return err
		}

		a.checkout.Path = path
	}

	// which git
	if a.checkout.GitPath == "" {
		gitBin, err := cmd(ctx, "", "which", "git")
		if err != nil {
			return err
		}

		if len(gitBin) == 0 {
			return fmt.Errorf("%w: git", ErrBinNotFound)
		}

		a.checkout.GitPath = gitBin[0]
	}

	path := buildPath(a.repo)

	var out []string
	var err error

	// git clone --depth=1 {remote} {directory}
	switch {
	case len(a.checkout.CommandOverrides) > 0:
		for i := range a.checkout.CommandOverrides {
			args := append(strings.Split(a.checkout.CommandOverrides[i], " "), a.checkout.Path)
			output, err := cmd(ctx, "", a.checkout.GitPath, args...)
			if err != nil {
				return err
			}

			out = append(out, output...)
		}

	default:
		out, err = cmd(ctx, "", a.checkout.GitPath, "clone", "--depth=1", path, a.checkout.Path)
		if err != nil {
			return err
		}
	}

	a.logger.InfoContext(ctx, "checked out repository",
		slog.Any("output", out),
	)

	// TODO: add branches support

	return nil
}

func Checkout(
	ctx context.Context,
	repo *config.Repository, cfg *config.Checkout,
	logger *slog.Logger,
) error {
	return (&ModUpdate{
		repo:     repo,
		checkout: cfg,
		logger:   logger,
	}).Checkout(ctx)
}

func buildPath(repo *config.Repository) string {
	sb := &strings.Builder{}
	sb.WriteString("https://")

	if repo.Username != "" {
		sb.WriteString(repo.Username)

		if repo.Token != "" {
			sb.WriteByte(':')
		}
	}

	if repo.Token != "" {
		sb.WriteString(repo.Token)
		sb.WriteByte('@')
	}

	sb.WriteString(repo.Path)

	return sb.String()
}
