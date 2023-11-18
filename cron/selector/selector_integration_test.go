//go:build integration

package selector_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/cron/selector"
	"github.com/zalgonoise/x/is"
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

func TestSelector(t *testing.T) {
	h := slog.NewJSONHandler(os.Stderr, nil)

	values := make(chan int)
	runner1 := testRunner{v: 1, ch: values}
	runner2 := testRunner{v: 2, ch: values}

	cron := "* * * * * *"
	twoMinEven := "0/2 * * * * *"
	twoMinOdd := "1/2 * * * * *"
	defaultDur := 2100 * time.Millisecond

	for _, testcase := range []struct {
		name    string
		execMap map[string][]executor.Runner // cron string : runners
		dur     time.Duration
		wants   []int
		err     error
	}{
		{
			name: "SingleExecTwoRunners",
			execMap: map[string][]executor.Runner{
				cron: {runner1, runner2},
			},
			dur:   defaultDur,
			wants: []int{1, 1, 2, 2},
		},
		{
			name: "TwoExecsTwoRunners",
			execMap: map[string][]executor.Runner{
				twoMinEven: {runner1},
				twoMinOdd:  {runner2},
			},
			dur:   defaultDur * 2,
			wants: []int{1, 1, 2, 2},
		},
		{
			name: "TwoExecsOffsetFrequency",
			execMap: map[string][]executor.Runner{
				cron:      {runner1},
				twoMinOdd: {runner2},
			},
			dur:   defaultDur,
			wants: []int{1, 1, 2},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			results := make([]int, 0, len(testcase.wants))
			execs := make([]executor.Executor, 0, len(testcase.execMap))

			var n int
			for cronString, runners := range testcase.execMap {
				exec, err := executor.New(fmt.Sprintf("%d", n),
					executor.WithSchedule(cronString),
					executor.WithLocation(time.Local),
					executor.WithRunners(runners...),
					executor.WithLogHandler(h),
				)
				is.Empty(t, err)

				execs = append(execs, exec)
				n++
			}

			sel, err := selector.New(
				selector.WithExecutors(execs...),
				selector.WithLogHandler(h),
			)

			is.Empty(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), testcase.dur)

			go func(t *testing.T, err error) {
				defer cancel()

				for {
					select {
					case <-ctx.Done():
						return
					default:
					}

					selErr := sel.Next(ctx)
					if !errors.Is(selErr, err) && !errors.Is(selErr, context.DeadlineExceeded) {
						t.Error(selErr)
					}
				}
			}(t, testcase.err)

			for {
				select {
				case <-ctx.Done():
					if testcase.dur < time.Second {
						is.True(t, errors.Is(ctx.Err(), context.DeadlineExceeded))

						return
					}

					slices.Sort(results)
					is.EqualElements(t, testcase.wants, results)

					return
				case v := <-values:
					t.Log("received", v)

					results = append(results, v)
				}
			}
		})
	}
}
