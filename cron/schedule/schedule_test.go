package schedule

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/zalgonoise/x/cron/log"
	"github.com/zalgonoise/x/cron/metrics"
	"github.com/zalgonoise/x/cron/schedule/cronlex"
	"github.com/zalgonoise/x/is"
	"go.opentelemetry.io/otel/trace"
)

func TestCronSchedule_Next(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		cronString string
		sched      Scheduler
		input      time.Time
		wants      time.Time
		err        error
	}{
		{
			name:       "Success/EverySecond",
			cronString: "* * * * * *",
			input:      time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 30, 10, 12, 44, 0, time.UTC),
		},
		{
			name:       "Success/EveryFifthSecond",
			cronString: "*/5 * * * * *",
			input:      time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 30, 10, 12, 45, 0, time.UTC),
		},
		{
			name:       "Success/EveryFifthSecondGoNext",
			cronString: "*/5 * * * * *",
			input:      time.Date(2023, 10, 30, 10, 12, 45, 0, time.UTC),
			wants:      time.Date(2023, 10, 30, 10, 12, 50, 0, time.UTC),
		},
		{
			name:       "Success/SecondsOddCombo",
			cronString: "0/3,2 * * * * *",

			input: time.Date(2023, 10, 30, 10, 12, 45, 0, time.UTC),
			wants: time.Date(2023, 10, 30, 10, 12, 48, 0, time.UTC),
		},
		{
			name:       "Success/EveryMinute",
			cronString: "* * * * *",
			input:      time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 30, 10, 13, 0, 0, time.UTC),
		},
		{
			name:       "Success/OneHour",
			cronString: "0 * * * *",
			input:      time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 30, 11, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/OneDay/WithDayChange",
			cronString: "0 0 * * *",
			input:      time.Date(2023, 10, 30, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 31, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithWeekday/NoWeekends",
			cronString: "0 0 * * 1-5",
			input:      time.Date(2023, 10, 30, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 31, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithWeekday/NoWeekendsAndWednesdays",
			cronString: "0 0 * * 1,2,4,5",
			input:      time.Date(2023, 10, 31, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 11, 2, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithRangesAndSteps/NoWeekendsAndWednesdays",
			cronString: "0 0 * * 1-2,4-5",
			input:      time.Date(2023, 10, 31, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 11, 2, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithRangesAndSteps/NoWeekendsAndWednesdays",
			cronString: "0 0/3,2 * * 1-2,4-5",
			input:      time.Date(2023, 10, 31, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 11, 2, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithWeekday/NoWeekendsStepSchedule",
			cronString: "0 0 * * 1,2,3,4,5",
			input:      time.Date(2023, 10, 30, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 31, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/WithStepSchedule/Every3Hours",
			cronString: "0 */3 * * *",
			input:      time.Date(2023, 10, 30, 22, 12, 43, 0, time.UTC),
			wants:      time.Date(2023, 10, 31, 00, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/EveryMinuteFromZeroToFive",
			cronString: "0-5 * * * *",

			input: time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants: time.Date(2023, 10, 30, 11, 0, 0, 0, time.UTC),
		},
		{
			name:       "Success/InvalidCronString",
			cronString: "*",

			input: time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			err:   cronlex.ErrInvalidNodeType,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sched, err := New(
				WithSchedule(testcase.cronString),
				WithLocation(time.UTC),
				WithLogHandler(log.NoOp()),
				WithMetrics(metrics.NoOp()),
				WithTrace(trace.NewNoopTracerProvider().Tracer("test")),
			)
			if testcase.err != nil {
				is.True(t, errors.Is(err, testcase.err))

				return
			}

			is.Empty(t, err)

			next := sched.Next(context.Background(), testcase.input)

			is.Equal(t, testcase.wants, next)
		})
	}
}

func TestConfig(t *testing.T) {
	t.Run("WithLogger", func(t *testing.T) {
		_, err := New(
			WithSchedule("* * * * *"),
			WithLocation(time.UTC),
			WithLogger(slog.New(log.NoOp())),
		)

		is.Empty(t, err)
	})

	t.Run("AllEmptyOptions", func(t *testing.T) {
		_, err := New(
			WithSchedule(""),
			WithLocation(nil),
			WithLogger(nil),
			WithLogHandler(nil),
			WithMetrics(nil),
			WithTrace(nil),
		)

		is.True(t, errors.Is(err, cronlex.ErrEmptyInput))
	})
}

func TestSchedulerWithLogs(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		s              Scheduler
		handler        slog.Handler
		wants          Scheduler
		defaultHandler bool
	}{
		{
			name:  "NilScheduler",
			wants: noOpScheduler{},
		},
		{
			name: "NilHandler",
			s:    noOpScheduler{},
			wants: withLogs{
				s: noOpScheduler{},
			},
			defaultHandler: true,
		},
		{
			name:    "WithHandler",
			s:       noOpScheduler{},
			handler: log.NoOp(),
			wants: withLogs{
				s:      noOpScheduler{},
				logger: slog.New(log.NoOp()),
			},
		},
		{
			name: "ReplaceHandler",
			s: withLogs{
				s: noOpScheduler{},
			},
			handler: log.NoOp(),
			wants: withLogs{
				s:      noOpScheduler{},
				logger: slog.New(log.NoOp()),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := schedulerWithLogs(testcase.s, testcase.handler)

			switch sched := s.(type) {
			case noOpScheduler:
				is.Equal(t, testcase.wants, s)
			case withLogs:
				wants, ok := testcase.wants.(withLogs)
				is.True(t, ok)

				is.Equal(t, wants.s, sched.s)
				if testcase.defaultHandler {
					is.True(t, sched.logger.Handler() != nil)

					return
				}

				is.Equal(t, wants.logger.Handler(), sched.logger.Handler())
			}
		})
	}
}

func TestSchedulerWithMetrics(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		s     Scheduler
		m     Metrics
		wants Scheduler
	}{
		{
			name:  "NilScheduler",
			wants: noOpScheduler{},
		},
		{
			name:  "NilMetrics",
			s:     noOpScheduler{},
			wants: noOpScheduler{},
		},
		{
			name: "WithMetrics",
			s:    noOpScheduler{},
			m:    metrics.NoOp(),
			wants: withMetrics{
				s: noOpScheduler{},
				m: metrics.NoOp(),
			},
		},
		{
			name: "ReplaceMetrics",
			s: withMetrics{
				s: noOpScheduler{},
			},
			m: metrics.NoOp(),
			wants: withMetrics{
				s: noOpScheduler{},
				m: metrics.NoOp(),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := schedulerWithMetrics(testcase.s, testcase.m)

			switch sched := s.(type) {
			case noOpScheduler:
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

func TestSchedulerWithTrace(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		s      Scheduler
		tracer trace.Tracer
		wants  Scheduler
	}{
		{
			name:  "NilScheduler",
			wants: noOpScheduler{},
		},
		{
			name:  "NilTracer",
			s:     noOpScheduler{},
			wants: noOpScheduler{},
		},
		{
			name:   "WithTracer",
			s:      noOpScheduler{},
			tracer: trace.NewNoopTracerProvider().Tracer("test"),
			wants: withTrace{
				s:      noOpScheduler{},
				tracer: trace.NewNoopTracerProvider().Tracer("test"),
			},
		},
		{
			name: "ReplaceTracer",
			s: withTrace{
				s: noOpScheduler{},
			},
			tracer: trace.NewNoopTracerProvider().Tracer("test"),
			wants: withTrace{
				s:      noOpScheduler{},
				tracer: trace.NewNoopTracerProvider().Tracer("test"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			s := schedulerWithTrace(testcase.s, testcase.tracer)

			switch sched := s.(type) {
			case noOpScheduler:
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
	noop := NoOp()

	is.Equal(t, time.Time{}, noop.Next(context.Background(), time.Now()))
}
