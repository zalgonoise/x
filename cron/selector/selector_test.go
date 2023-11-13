package selector

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
	runner := executor.Runnable(func(context.Context) error {
		return nil
	})
	cronString := "* * * * * *"

	exec, err := executor.New("test",
		executor.WithRunners(runner),
		executor.WithSchedule(cronString),
	)
	is.Empty(t, err)

	for _, testcase := range []struct {
		name string
		opts []cfg.Option[Config]
	}{
		{
			name: "WithExecutors/NoExecutors",
			opts: []cfg.Option[Config]{
				WithExecutors(),
			},
		},
		{
			name: "WithExecutors/MultipleCalls",
			opts: []cfg.Option[Config]{
				WithExecutors(exec),
				WithExecutors(exec),
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

func TestSelectorWithLogs(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		s              Selector
		handler        slog.Handler
		wants          Selector
		defaultHandler bool
	}{
		{
			name:  "NilSelector",
			wants: noOpSelector{},
		},
		{
			name: "NilHandler",
			s:    noOpSelector{},
			wants: withLogs{
				s: noOpSelector{},
			},
			defaultHandler: true,
		},
		{
			name:    "WithHandler",
			s:       noOpSelector{},
			handler: log.NoOp(),
			wants: withLogs{
				s:      noOpSelector{},
				logger: slog.New(log.NoOp()),
			},
		},
		{
			name: "ReplaceHandler",
			s: withLogs{
				s: noOpSelector{},
			},
			handler: log.NoOp(),
			wants: withLogs{
				s:      noOpSelector{},
				logger: slog.New(log.NoOp()),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := selectorWithLogs(testcase.s, testcase.handler)

			_ = s.Next(context.Background())

			switch exec := s.(type) {
			case noOpSelector:
				is.Equal(t, testcase.wants, s)
			case withLogs:
				wants, ok := testcase.wants.(withLogs)
				is.True(t, ok)

				is.Equal(t, wants.s, exec.s)
				if testcase.defaultHandler {
					is.True(t, exec.logger.Handler() != nil)

					return
				}

				is.Equal(t, wants.logger.Handler(), exec.logger.Handler())
			}
		})
	}
}

func TestSelectorWithMetrics(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		s     Selector
		m     Metrics
		wants Selector
	}{
		{
			name:  "NilSelector",
			wants: noOpSelector{},
		},
		{
			name:  "NilMetrics",
			s:     noOpSelector{},
			wants: noOpSelector{},
		},
		{
			name: "WithMetrics",
			s:    noOpSelector{},
			m:    metrics.NoOp(),
			wants: withMetrics{
				s: noOpSelector{},
				m: metrics.NoOp(),
			},
		},
		{
			name: "ReplaceMetrics",
			s: withMetrics{
				s: noOpSelector{},
			},
			m: metrics.NoOp(),
			wants: withMetrics{
				s: noOpSelector{},
				m: metrics.NoOp(),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := selectorWithMetrics(testcase.s, testcase.m)

			_ = s.Next(context.Background())

			switch sched := s.(type) {
			case noOpSelector:
				is.Equal(t, testcase.wants, s)
			case withMetrics:
				wants, ok := testcase.wants.(withMetrics)
				is.True(t, ok)
				is.Equal(t, wants.s, sched.s)
				is.Equal(t, wants.m, sched.m)
			}
		})
	}
}

func TestSelectorWithTrace(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		s      Selector
		tracer trace.Tracer
		wants  Selector
	}{
		{
			name:  "NilSelector",
			wants: noOpSelector{},
		},
		{
			name:  "NilTracer",
			s:     noOpSelector{},
			wants: noOpSelector{},
		},
		{
			name:   "WithTracer",
			s:      noOpSelector{},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				s:      noOpSelector{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
		{
			name: "ReplaceTracer",
			s: withTrace{
				s: noOpSelector{},
			},
			tracer: noop.NewTracerProvider().Tracer("test"),
			wants: withTrace{
				s:      noOpSelector{},
				tracer: noop.NewTracerProvider().Tracer("test"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := selectorWithTrace(testcase.s, testcase.tracer)

			_ = s.Next(context.Background())

			switch sched := s.(type) {
			case noOpSelector:
				is.Equal(t, testcase.wants, s)
			case withTrace:
				wants, ok := testcase.wants.(withTrace)
				is.True(t, ok)
				is.Equal(t, wants.s, sched.s)
				is.Equal(t, wants.tracer, sched.tracer)
			}
		})
	}
}

func TestNoOp(t *testing.T) {
	noOp := NoOp()

	is.Empty(t, noOp.Next(context.Background()))
}

func TestWithObservability(t *testing.T) {
	runner := executor.Runnable(func(context.Context) error {
		return nil
	})

	testErr := errors.New("test error")
	errRunner := executor.Runnable(func(context.Context) error {
		return testErr
	})
	cronString := "* * * * * *"

	for _, testcase := range []struct {
		name   string
		runner executor.Runner
		err    error
	}{
		{
			name:   "Success",
			runner: runner,
		},

		{
			name:   "WithError",
			runner: errRunner,
			err:    testErr,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			exec, err := executor.New("test",
				executor.WithRunners(testcase.runner),
				executor.WithSchedule(cronString),
			)
			is.Empty(t, err)

			sel, err := New(
				WithExecutors(exec),
				WithLogHandler(log.NoOp()),
				WithLogger(slog.New(log.NoOp())),
				WithMetrics(metrics.NoOp()),
				WithTrace(noop.NewTracerProvider().Tracer("test")),
			)
			is.Empty(t, err)

			err = sel.Next(context.Background())
			is.True(t, errors.Is(err, testcase.err))
		})
	}
}

func TestZeroExecutors(t *testing.T) {
	t.Run("FromRawSelector", func(t *testing.T) {
		is.True(t, errors.Is(
			ErrEmptyExecutorsList,
			selector{}.Next(context.Background()),
		))
	})

	t.Run("FromConstructor", func(t *testing.T) {
		_, err := New()
		is.True(t, errors.Is(ErrEmptyExecutorsList, err))
	})
}
