//go:build integration

package executor_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/cron/log"
	"github.com/zalgonoise/x/cron/metrics"
	"github.com/zalgonoise/x/is"
	"go.opentelemetry.io/otel/trace"
)

type testRunner struct {
	v  int
	ch chan<- int

	err error
}

func (r testRunner) Run(context.Context) error {
	r.ch <- r.v

	return r.err
}

func TestExecutor(t *testing.T) {
	testErr := errors.New("test error")
	values := make(chan int)
	runner1 := testRunner{v: 1, ch: values}
	runner2 := testRunner{v: 2, ch: values}
	runner3 := testRunner{v: 3, ch: values, err: testErr}
	cron := "* * * * *"
	defaultDur := 70 * time.Second

	for _, testcase := range []struct {
		name    string
		dur     time.Duration
		runners []executor.Runner
		wants   []int
		err     error
	}{
		{
			name:    "ContextCanceled",
			dur:     10 * time.Millisecond,
			runners: []executor.Runner{runner1, runner2},
			err:     context.DeadlineExceeded,
		},
		{
			name:    "TwoRunners",
			dur:     defaultDur,
			runners: []executor.Runner{runner1, runner2},
			wants:   []int{1, 2},
		},
		{
			name:    "ErrorRunner",
			dur:     defaultDur,
			runners: []executor.Runner{runner3},
			wants:   []int{3},
			err:     testErr,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {

			exec, err := executor.New(testcase.name,
				executor.WithSchedule(cron),
				executor.WithLocation(time.Local),
				executor.WithRunners(testcase.runners...),
				executor.WithMetrics(metrics.NoOp()),
				executor.WithLogHandler(log.NoOp()),
				executor.WithLogger(slog.New(log.NoOp())),
				executor.WithTrace(trace.NewNoopTracerProvider().Tracer("test")),
			)
			is.Empty(t, err)
			is.Equal(t, testcase.name, exec.ID())

			results := make([]int, 0, len(testcase.wants))

			// run test for 1min 10sec
			ctx, cancel := context.WithTimeout(context.Background(), testcase.dur)
			defer cancel()

			go func() {
				err = exec.Exec(ctx)
				if err != nil {
					is.True(t, errors.Is(err, testcase.err))
				}

				cancel()
			}()

			for {
				select {
				case <-ctx.Done():
					if testcase.dur < time.Second {
						is.True(t, errors.Is(ctx.Err(), context.DeadlineExceeded))

						return
					}

					is.EqualElements(t, testcase.wants, results)

					return
				case v := <-values:
					results = append(results, v)
				}
			}
		})
	}
}
