package actions

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/zalgonoise/x/modupdate/events"
)

const (
	defaultBranch        = "main"
	defaultCommitMessage = "chore: updated modules"
)

func (a *ModUpdate) Push(ctx context.Context) error {
	dir := a.checkout.Path
	if a.repo.ModulePath != "" {
		dir = path.Join(dir, a.repo.ModulePath)
	}

	// which git
	if err := a.setGit(ctx); err != nil {
		return err
	}

	// git add go.mod go.sum
	// git commit -m 'chore: updated modules'
	out, err := a.gitAddGoMod(ctx, dir)
	if err != nil {
		return err
	}

	if a.push.DryRun {
		return a.doDryRun(ctx, dir, out)
	}

	return a.gitPush(ctx, dir, out)
}

func (a *ModUpdate) gitAddGoMod(ctx context.Context, dir string) (out []string, err error) {
	switch {
	case len(a.push.CommandOverrides) > 0:
		for i := range a.push.CommandOverrides {
			args := append(strings.Split(a.push.CommandOverrides[i], " "), a.checkout.Path)
			output, err := cmd(ctx, dir, a.checkout.GitPath, args...)
			if err != nil {
				if len(output) > 0 {
					err = fmt.Errorf("%w: %s", err, strings.Join(output, "; "))
				}

				return nil, err
			}

			out = append(out, output...)
		}

	default:
		output, err := cmd(ctx, dir, a.checkout.GitPath, "add", "go.mod", "go.sum")
		if err != nil {
			if len(output) > 0 {
				err = fmt.Errorf("%w: %s", err, strings.Join(output, "; "))
			}

			return nil, err
		}

		out = append(out, output...)

		if a.push.CommitMessage == "" {
			a.push.CommitMessage = defaultCommitMessage
		}

		output, err = cmd(ctx, dir, a.checkout.GitPath, "commit", "-m", a.push.CommitMessage)
		if err != nil {
			return out, err
		}

		out = append(out, output...)
	}

	return out, nil
}

func (a *ModUpdate) doDryRun(ctx context.Context, dir string, out []string) error {
	output, err := cmd(ctx, dir, a.checkout.GitPath, "status")
	if err != nil {
		return err
	}

	out = append(out, output...)

	a.reporter.ReportEvent(ctx, events.Event{
		Action: actionPushCommit,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	a.logger.InfoContext(ctx, "files committed successfully",
		slog.Bool("dry_run", true), slog.Any("output", out))

	return nil
}

func (a *ModUpdate) gitPush(ctx context.Context, dir string, out []string) error {
	targetBranch := a.repo.Branch
	if targetBranch == "" {
		targetBranch = defaultBranch
	}

	// git push
	output, err := cmd(ctx, dir, a.checkout.GitPath, "push", "-u", buildPath(a.repo), targetBranch)
	if err != nil {
		return err
	}

	out = append(out, output...)

	a.reporter.ReportEvent(ctx, events.Event{
		Action: actionPushPush,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})
	a.logger.InfoContext(ctx, "files pushed to origin",
		slog.Bool("dry_run", true), slog.Any("output", out))

	return nil
}
