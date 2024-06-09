package actions

import (
	"context"
	"fmt"
	"log/slog"
	"path"
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

	out := make([]string, 0, 8)

	// git add go.mod go.sum
	// git commit -m 'chore: updated modules'
	switch {
	case len(a.push.CommandOverrides) > 0:
	default:
		output, err := cmd(ctx, dir, a.checkout.GitPath, "add", "go.mod", "go.sum")
		if err != nil {
			return err
		}

		out = append(out, output...)

		if a.push.CommitMessage == "" {
			a.push.CommitMessage = defaultCommitMessage
		}

		output, err = cmd(ctx, dir, a.checkout.GitPath, "commit", "-m", a.push.CommitMessage)
		if err != nil {
			return err
		}

		out = append(out, output...)
	}

	if a.push.DryRun {
		output, err := cmd(ctx, dir, a.checkout.GitPath, "status")
		if err != nil {
			return err
		}

		out = append(out, output...)

		a.logger.InfoContext(ctx, "files committed successfully",
			slog.Bool("dry_run", true), slog.Any("output", out))

		return nil
	}

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

	a.logger.InfoContext(ctx, "files pushed to origin",
		slog.Bool("dry_run", true), slog.Any("output", out))

	return nil
}
