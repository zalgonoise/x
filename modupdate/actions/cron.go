package actions

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/micron"
	"github.com/zalgonoise/micron/executor"
	"github.com/zalgonoise/micron/selector"
	"github.com/zalgonoise/x/modupdate/config"
)

const serviceID = "modupdate"

func NewActions(reporter Reporter, logger *slog.Logger, tasks ...*config.Task) (micron.Runtime, error) {
	execs := make([]executor.Executor, 0, len(tasks))

	for i := range tasks {
		e, err := executor.New(serviceID,
			executor.WithSchedule(tasks[i].CronSchedule),
			executor.WithRunners(NewModUpdate(reporter, tasks[i], logger)),
			executor.WithLocation(time.Local),
			executor.WithLogger(logger),
		)

		if err != nil {
			return nil, err
		}

		execs = append(execs, e)
	}

	sel, err := selector.New(selector.WithExecutors(execs...))
	if err != nil {
		return nil, err
	}

	return micron.New(
		micron.WithLogger(logger),
		micron.WithSelector(sel),
	)
}
