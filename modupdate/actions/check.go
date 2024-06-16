package actions

import (
	"context"
	"log/slog"
	"path"
	"strings"

	"github.com/zalgonoise/x/modupdate/events"
)

func (a *ModUpdate) Check(ctx context.Context) error {
	if a.check.Skip {
		return nil
	}

	dir := a.checkout.Path
	if a.repo.ModulePath != "" {
		dir = path.Join(dir, a.repo.ModulePath)
	}

	goBin := parseGoBin(a.update.GoBin)
	if goBin == "" {
		goBin = "go"
	}

	return a.checkGoBuild(ctx, dir, goBin)
}

func (a *ModUpdate) checkGoBuild(ctx context.Context, dir, goBin string) error {
	out := make([]string, 0, 8)

	switch {
	case len(a.check.CommandOverrides) > 0:
		for i := range a.check.CommandOverrides {
			output, err := cmd(ctx, dir, goBin, strings.Split(a.check.CommandOverrides[i], " ")...)
			if err != nil {
				return err
			}

			out = append(out, output...)
		}
	default:
		output, err := cmd(ctx, dir, goBin, "build", "-o", "/dev/null", "./...")
		if err != nil {
			return err
		}

		out = append(out, output...)
	}

	a.reporter.ReportEvent(ctx, events.Event{
		Action: actionCheckBuild,
		URI:    a.repo.Path,
		Module: a.repo.ModulePath,
		Branch: a.repo.Branch,
		Output: out,
	})

	a.logger.InfoContext(ctx, "module builds successfully", slog.Any("output", out))

	return nil
}
