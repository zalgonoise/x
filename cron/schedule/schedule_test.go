package schedule

import (
	"context"
	"errors"
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
			name:       "Success/EveryMinute",
			cronString: "* * * * *",

			input: time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants: time.Date(2023, 10, 30, 10, 13, 0, 0, time.UTC),
		},
		{
			name:       "Success/OneHour",
			cronString: "0 * * * *",

			input: time.Date(2023, 10, 30, 10, 12, 43, 0, time.UTC),
			wants: time.Date(2023, 10, 30, 11, 0, 0, 0, time.UTC),
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

func TestNoOp(t *testing.T) {
	noop := NoOp()

	is.Equal(t, time.Time{}, noop.Next(context.Background(), time.Now()))
}
