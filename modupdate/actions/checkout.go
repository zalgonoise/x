package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zalgonoise/x/modupdate/config"
	"github.com/zalgonoise/x/modupdate/events"
)

var ErrEmptyPath = errors.New("target path to checkout into is empty")

func (a *ModUpdate) Checkout(ctx context.Context) error {
	if err := a.setPath(ctx); err != nil {
		return err
	}

	// check if repo has already been checked out
	ok, err := a.checkRepoExists(ctx)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	// which git
	if err := a.setGit(ctx); err != nil {
		return err
	}

	// git clone --depth=1 {remote} {directory}
	return a.gitClone(ctx)
}

func (a *ModUpdate) setPath(ctx context.Context) error {
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

	return nil
}

func (a *ModUpdate) checkRepoExists(ctx context.Context) (bool, error) {
	if out, err := cmd(ctx, a.checkout.Path, "file", ".git"); err == nil {
		a.reporter.ReportEvent(ctx, events.Event{
			Action: actionCheckoutPresent,
			URI:    a.repo.Path,
			Module: a.repo.ModulePath,
			Branch: a.repo.Branch,
			Output: out,
		})

		a.logger.InfoContext(ctx, "repository has already been checked out",
			slog.Any("output", out),
		)

		return true, a.checkoutBranch(ctx)
	}

	return false, nil
}

func (a *ModUpdate) setGit(ctx context.Context) error {
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

	return nil
}

func (a *ModUpdate) gitClone(ctx context.Context) (err error) {
	var out []string
	path := buildPath(a.repo)

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

	a.reporter.ReportEvent(ctx, events.Event{
		Action: actionCheckout,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	a.logger.InfoContext(ctx, "checked out repository",
		slog.Any("output", out),
	)

	return a.checkoutBranch(ctx)
}

func (a *ModUpdate) checkoutBranch(ctx context.Context) error {
	out, err := cmd(ctx, a.checkout.Path, "git", "checkout", a.repo.Branch)
	if err != nil {
		return err
	}

	a.reporter.ReportEvent(ctx, events.Event{
		Action: actionCheckoutBranch,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	a.logger.InfoContext(ctx, "checked out repository's branch",
		slog.String("branch", a.repo.Branch),
		slog.Any("output", out),
	)

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
