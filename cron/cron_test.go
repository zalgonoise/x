package cron

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/cron/log"
	"github.com/zalgonoise/x/cron/metrics"
	"github.com/zalgonoise/x/is"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestConfig(t *testing.T) {
	for _, testcase := range []struct {
		name string
		opts []cfg.Option[Config]
	}{
		{
			name: "WithErrorBufferSize/Zero",
			opts: []cfg.Option[Config]{
				WithErrorBufferSize(0),
			},
		},
		{
			name: "WithErrorBufferSize/Ten",
			opts: []cfg.Option[Config]{
				WithErrorBufferSize(10),
			},
		},
		{
			name: "WithErrorBufferSize/Negative",
			opts: []cfg.Option[Config]{
				WithErrorBufferSize(-10),
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

func TestRuntimeWithLogs(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		r              Runtime
		handler        slog.Handler
		wants          Runtime
		defaultHandler bool
	}{
		{
			name:  "NilRuntime",
			wants: noOpRuntime{},
		},
		{
			name: "NilHandler",
			r:    noOpRuntime{},
			wants: withLogs{
				r: noOpRuntime{},
			},
			defaultHandler: true,
		},
		{
			name:    "WithHandler",
			r:       noOpRuntime{},
			handler: log.NoOp(),
			wants: withLogs{
				r:      noOpRuntime{},
				logger: slog.New(log.NoOp()),
			},
		},
		{
			name: "ReplaceHandler",
			r: withLogs{
				r: noOpRuntime{},
			},
			handler: log.NoOp(),
			wants: withLogs{
				r:      noOpRuntime{},
				logger: slog.New(log.NoOp()),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			r := cronWithLogs(testcase.r, testcase.handler)

			r.Run(context.Background())
			_ = r.Err()

			switch exec := r.(type) {
			case noOpRuntime:
				is.Equal(t, testcase.wants, r)
			case withLogs:
				wants, ok := testcase.wants.(withLogs)
				is.True(t, ok)

				is.Equal(t, wants.r, exec.r)
				if testcase.defaultHandler {
					is.True(t, exec.logger.Handler() != nil)

					return
				}

				is.Equal(t, wants.logger.Handler(), exec.logger.Handler())
			}
		})
	}
}

func TestRuntimeWithMetrics(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		r     Runtime
		m     Metrics
		wants Runtime
	}{
		{
			name:  "NilRuntime",
			wants: noOpRuntime{},
		},
		{
			name:  "NilMetrics",
			r:     noOpRuntime{},
			wants: noOpRuntime{},
		},
		{
			name: "WithMetrics",
			r:    noOpRuntime{},
			m:    metrics.NoOp(),
			wants: withMetrics{
				r: noOpRuntime{},
				m: metrics.NoOp(),
			},
		},
		{
			name: "ReplaceMetrics",
			r: withMetrics{
				r: noOpRuntime{},
			},
			m: metrics.NoOp(),
			wants: withMetrics{
				r: noOpRuntime{},
				m: metrics.NoOp(),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			r := cronWithMetrics(testcase.r, testcase.m)

			r.Run(context.Background())
			_ = r.Err()

			switch sched := r.(type) {
			case noOpRuntime:
				is.Equal(t, testcase.wants, r)
			case withMetrics:
				wants, ok := testcase.wants.(withMetrics)
				is.True(t, ok)
				is.Equal(t, wants.r, sched.r)
				is.Equal(t, wants.m, sched.m)
			}
		})
	}
}

func TestRuntimeWithTrace(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		r      Runtime
		tracer trace.Tracer
		wants  Runtime
	}{
		{
			name:  "NilRuntime",
			wants: noOpRuntime{},
		},
		{
			name:  "NilTracer",
			r:     noOpRuntime{},
			wants: noOpRuntime{},
		},
		{
			name:   "WithTracer",
			r:      noOpRuntime{},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				r:      noOpRuntime{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
		{
			name: "ReplaceTracer",
			r: withTrace{
				r: noOpRuntime{},
			},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				r:      noOpRuntime{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			r := cronWithTrace(testcase.r, testcase.tracer)

			r.Run(context.Background())
			_ = r.Err()

			switch sched := r.(type) {
			case noOpRuntime:
				is.Equal(t, testcase.wants, r)
			case withTrace:
				wants, ok := testcase.wants.(withTrace)
				is.True(t, ok)
				is.Equal(t, wants.r, sched.r)
				is.Equal(t, wants.tracer, sched.tracer)
			}
		})
	}
}

func TestNoOp(t *testing.T) {
	noOp := NoOp()

	noOp.Run(context.Background())
	is.Empty(t, noOp.Err())
}

func TestNew_NilSelector(t *testing.T) {
	_, err := New(nil)
	is.True(t, errors.Is(err, ErrEmptySelector))
}

func TestNewWithJob(t *testing.T) {
	runner1 := executor.Runnable(func(ctx context.Context) error {
		return nil
	})
	runner2 := executor.Runnable(func(ctx context.Context) error {
		return nil
	})

	type job struct {
		id         string
		cronString string
		runners    []executor.Runner
	}

	for _, testcase := range []struct {
		name string
		jobs []job
		err  error
	}{
		{
			name: "Success/SingleRunner",
			jobs: []job{{
				id:         "ok-job",
				cronString: "* * * * * *",
				runners:    []executor.Runner{runner1},
			}},
		},
		{
			name: "Success/MultiRunner",
			jobs: []job{{
				id:         "ok-job",
				cronString: "* * * * * *",
				runners:    []executor.Runner{runner1, runner2},
			}},
		},
		{
			name: "Success/NoID/MultiRunner",
			jobs: []job{{
				cronString: "* * * * * *",
				runners:    []executor.Runner{runner1, runner2},
			}},
		},
		{
			name: "Success/MultiJob/MultiRunner",
			jobs: []job{
				{
					id:         "seconds",
					cronString: "* * * * * *",
					runners:    []executor.Runner{runner1},
				},
				{
					id:         "minutes",
					cronString: "* * * * *",
					runners:    []executor.Runner{runner2},
				},
			},
		},
		{
			name: "Fail/NoCronString",
			jobs: []job{{
				id:      "ok-job",
				runners: []executor.Runner{runner1, runner2},
			}},
			err: ErrEmptySelector,
		},
		{
			name: "Fail/NoRunners",
			jobs: []job{{
				id:         "ok-job",
				cronString: "* * * * * *",
			}},
			err: ErrEmptySelector,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			jobs := make([]cfg.Option[Config], 0, len(testcase.jobs))

			for i := range testcase.jobs {
				jobs = append(jobs, WithJob(
					testcase.jobs[i].id,
					testcase.jobs[i].cronString,
					testcase.jobs[i].runners...,
				))
			}

			_, err := New(jobs...)
			t.Log(err)
			is.True(t, errors.Is(err, testcase.err))
		})
	}
}
