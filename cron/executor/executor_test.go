package executor

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/log"
	"github.com/zalgonoise/x/cron/metrics"
	"go.opentelemetry.io/otel/trace"
)

type testScheduler struct{}

func (testScheduler) Next(context.Context, time.Time) time.Time { return time.Time{} }

func TestOptions(t *testing.T) {
	runner := Runnable(func(context.Context) error {
		return nil
	})
	cron := "* * * * *"

	for _, testcase := range []struct {
		name string
		opts []cfg.Option[Config]
	}{
		{
			name: "WithRunners/NoRunners",
			opts: []cfg.Option[Config]{
				WithRunners(),
			},
		},
		{
			name: "WithRunners/OneRunner",
			opts: []cfg.Option[Config]{
				WithRunners(runner),
			},
		},
		{
			name: "WithRunners/AddRunner",
			opts: []cfg.Option[Config]{
				WithRunners(runner),
				WithRunners(runner),
			},
		},
		{
			name: "WithScheduler/NoScheduler",
			opts: []cfg.Option[Config]{
				WithScheduler(nil),
			},
		},
		{
			name: "WithScheduler/OneScheduler",
			opts: []cfg.Option[Config]{
				WithScheduler(testScheduler{}),
			},
		},
		{
			name: "WithSchedule/EmptyString",
			opts: []cfg.Option[Config]{
				WithSchedule(""),
			},
		},
		{
			name: "WithSchedule/WithCronString",
			opts: []cfg.Option[Config]{
				WithSchedule(cron),
			},
		},
		{
			name: "WithLocation/NilLocation",
			opts: []cfg.Option[Config]{
				WithLocation(nil),
			},
		},
		{
			name: "WithLocation/Local",
			opts: []cfg.Option[Config]{
				WithLocation(time.Local),
			},
		},
		{
			name: "WithMetrics/NilMetrics",
			opts: []cfg.Option[Config]{
				WithMetrics(nil),
			},
		},
		{
			name: "WithMetrics/NoOp",
			opts: []cfg.Option[Config]{
				WithMetrics(metrics.NoOp()),
			},
		},
		{
			name: "WithLogger/NilLogger",
			opts: []cfg.Option[Config]{
				WithLogger(nil),
			},
		},
		{
			name: "WithLogger/NoOp",
			opts: []cfg.Option[Config]{
				WithLogger(slog.New(log.NoOp())),
			},
		},
		{
			name: "WithLogHandler/NilHandler",
			opts: []cfg.Option[Config]{
				WithLogHandler(nil),
			},
		},
		{
			name: "WithLogHandler/NoOp",
			opts: []cfg.Option[Config]{
				WithLogHandler(log.NoOp()),
			},
		},
		{
			name: "WithTrace/NilTracer",
			opts: []cfg.Option[Config]{
				WithTrace(nil),
			},
		},
		{
			name: "WithTrace/NoOp",
			opts: []cfg.Option[Config]{
				WithTrace(trace.NewNoopTracerProvider().Tracer("test")),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_ = cfg.New(testcase.opts...)
		})
	}
}
