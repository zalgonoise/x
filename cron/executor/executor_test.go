package executor

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/log"
	"github.com/zalgonoise/x/cron/metrics"
	"github.com/zalgonoise/x/cron/schedule/cronlex"
	"github.com/zalgonoise/x/is"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type testScheduler struct{}

func (testScheduler) Next(context.Context, time.Time) time.Time { return time.Time{} }

func TestConfig(t *testing.T) {
	runner := Runnable(func(context.Context) error {
		return nil
	})
	cron := "* * * * * *"

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
				WithTrace(noop.NewTracerProvider().Tracer("test")),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_ = cfg.New(testcase.opts...)
		})
	}
}

func TestExecutorWithLogs(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		e              Executor
		handler        slog.Handler
		wants          Executor
		defaultHandler bool
	}{
		{
			name:  "NilExecutor",
			wants: noOpExecutor{},
		},
		{
			name: "NilHandler",
			e:    noOpExecutor{},
			wants: withLogs{
				e: noOpExecutor{},
			},
			defaultHandler: true,
		},
		{
			name:    "WithHandler",
			e:       noOpExecutor{},
			handler: log.NoOp(),
			wants: withLogs{
				e:      noOpExecutor{},
				logger: slog.New(log.NoOp()),
			},
		},
		{
			name: "ReplaceHandler",
			e: withLogs{
				e: noOpExecutor{},
			},
			handler: log.NoOp(),
			wants: withLogs{
				e:      noOpExecutor{},
				logger: slog.New(log.NoOp()),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			e := executorWithLogs(testcase.e, testcase.handler)

			switch exec := e.(type) {
			case noOpExecutor:
				is.Equal(t, testcase.wants, e)
			case withLogs:
				wants, ok := testcase.wants.(withLogs)
				is.True(t, ok)

				is.Equal(t, wants.e, exec.e)
				if testcase.defaultHandler {
					is.True(t, exec.logger.Handler() != nil)

					return
				}

				is.Equal(t, wants.logger.Handler(), exec.logger.Handler())

			}
		})
	}
}

func TestExecutorWithMetrics(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		e     Executor
		m     Metrics
		wants Executor
	}{
		{
			name:  "NilExecutor",
			wants: noOpExecutor{},
		},
		{
			name:  "NilMetrics",
			e:     noOpExecutor{},
			wants: noOpExecutor{},
		},
		{
			name: "WithMetrics",
			e:    noOpExecutor{},
			m:    metrics.NoOp(),
			wants: withMetrics{
				e: noOpExecutor{},
				m: metrics.NoOp(),
			},
		},
		{
			name: "ReplaceMetrics",
			e: withMetrics{
				e: noOpExecutor{},
			},
			m: metrics.NoOp(),
			wants: withMetrics{
				e: noOpExecutor{},
				m: metrics.NoOp(),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			e := executorWithMetrics(testcase.e, testcase.m)

			switch sched := e.(type) {
			case noOpExecutor:
				is.Equal(t, testcase.wants, e)
			case withMetrics:
				wants, ok := testcase.wants.(withMetrics)
				is.True(t, ok)
				is.Equal(t, wants.e, sched.e)
				is.Equal(t, wants.m, sched.m)
			}
		})
	}
}

func TestExecutorWithTrace(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		e      Executor
		tracer trace.Tracer
		wants  Executor
	}{
		{
			name:  "NilExecutor",
			wants: noOpExecutor{},
		},
		{
			name:  "NilTracer",
			e:     noOpExecutor{},
			wants: noOpExecutor{},
		},
		{
			name:   "WithTracer",
			e:      noOpExecutor{},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				e:      noOpExecutor{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
		{
			name: "ReplaceTracer",
			e: withTrace{
				e: noOpExecutor{},
			},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				e:      noOpExecutor{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			e := executorWithTrace(testcase.e, testcase.tracer)

			switch sched := e.(type) {
			case noOpExecutor:
				is.Equal(t, testcase.wants, e)
			case withTrace:
				wants, ok := testcase.wants.(withTrace)
				is.True(t, ok)
				is.Equal(t, wants.e, sched.e)
				is.Equal(t, wants.tracer, sched.tracer)
			}
		})
	}
}

func TestNoOp(t *testing.T) {
	noOp := NoOp()

	is.Equal(t, time.Time{}, noOp.Next(context.Background()))
	is.Equal(t, "", noOp.ID())
	is.Empty(t, noOp.Exec(context.Background()))
}

func TestNew(t *testing.T) {
	cronString := "@nope"
	r := Runnable(func(ctx context.Context) error {
		return nil
	})

	for _, testcase := range []struct {
		name string
		conf []cfg.Option[Config]
		err  error
	}{
		{
			name: "NoRunners",
			err:  ErrEmptyRunnerList,
		},
		{
			name: "NoSchedulerOrCronString",
			conf: []cfg.Option[Config]{
				WithRunners(r),
			},
			err: ErrEmptyScheduler,
		},
		{
			name: "InvalidCronString",
			conf: []cfg.Option[Config]{
				WithRunners(r),
				WithSchedule(cronString),
			},
			err: cronlex.ErrInvalidFrequency,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_, err := New(testcase.name, testcase.conf...)
			is.True(t, errors.Is(err, testcase.err))
		})
	}
}