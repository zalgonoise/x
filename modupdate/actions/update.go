package actions

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/zalgonoise/x/modupdate/config"
	"github.com/zalgonoise/x/modupdate/events"
)

func (a *ModUpdate) Update(ctx context.Context) error {
	dir := a.checkout.Path
	if a.repo.ModulePath != "" {
		dir = path.Join(dir, a.repo.ModulePath)
	}

	// which git
	if err := a.setGit(ctx); err != nil {
		return err
	}

	// git fetch
	// git pull --ff-only
	if err := a.gitFetchPull(ctx, dir); err != nil {
		return err
	}

	goBin := parseGoBin(a.update.GoBin)
	if goBin == "" {
		goBin = "go"
	}

	// go get -u ./...
	return a.goGet(ctx, dir, goBin)
}

func (a *ModUpdate) gitFetchPull(ctx context.Context, dir string) (err error) {
	out := make([]string, 0, 8)

	switch {
	case len(a.update.GitCommandOverrides) > 0:
		for i := range a.update.GitCommandOverrides {
			output, err := cmd(ctx, dir, a.checkout.GitPath, strings.Split(a.update.GitCommandOverrides[i], " ")...)
			if err != nil {
				if len(output) > 0 {
					err = fmt.Errorf("%w: %s", err, strings.Join(output, "; "))
				}

				return err
			}

			out = append(out, output...)
		}

	default:
		output, err := cmd(ctx, dir, a.checkout.GitPath, "fetch")
		if err != nil {
			if len(output) > 0 {
				err = fmt.Errorf("%w: %s", err, strings.Join(output, "; "))
			}

			return err
		}

		out = append(out, output...)

		output, err = cmd(ctx, dir, a.checkout.GitPath, "pull", "--ff-only")
		if err != nil {
			if len(output) > 0 {
				err = fmt.Errorf("%w: %s", err, strings.Join(output, "; "))
			}

			return err
		}

		out = append(out, output...)
	}

	a.logger.InfoContext(ctx, "repository updated successfully", slog.Any("output", out))

	a.reporter.ReportEvent(events.Event{
		Action: actionUpdateRepo,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	return nil
}

func (a *ModUpdate) goGet(ctx context.Context, dir, goBin string) error {
	out := make([]string, 0, 8)

	switch {
	case len(a.update.GoCommandOverrides) > 0:
		for i := range a.update.GoCommandOverrides {
			output, err := cmd(ctx, dir, goBin, strings.Split(a.update.GoCommandOverrides[i], " ")...)
			if err != nil {
				return err
			}

			out = append(out, output...)
		}
	default:
		output, err := cmd(ctx, dir, goBin, "get", "-u", "./...")
		if err != nil {
			return err
		}

		out = append(out, output...)
	}

	// go mod tidy
	tidyOut, err := cmd(ctx, dir, goBin, "mod", "tidy")
	if err != nil {
		return err
	}

	out = append(out, tidyOut...)

	a.reporter.ReportEvent(events.Event{
		Action: actionUpdateMod,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	a.logger.InfoContext(ctx, "modules updated successfully", slog.Any("output", out))

	return nil
}

func Update(
	ctx context.Context, dir string,
	repo *config.Repository, cfg *config.Update,
	logger *slog.Logger,
) error {
	return (&ModUpdate{
		repo: repo,
		checkout: &config.Checkout{
			Path: dir,
		},
		update: cfg,
		logger: logger,
	}).Update(ctx)
}

func parseGoBin(goBin string) string {
	if goBin == "" {
		return ""
	}

	home := os.Getenv("HOME")

	goBin = strings.Replace(goBin, "${HOME}", home, 1)
	goBin = strings.Replace(goBin, "$HOME", home, 1)
	goBin = strings.Replace(goBin, "~", home, 1)

	return goBin
}

func getPath(ctx context.Context, binary string) string {
	out, err := exec.CommandContext(ctx, "which", binary).CombinedOutput()
	if err != nil {
		return ""
	}

	// trim trailing newline if present
	if out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}

	return string(out)
}
