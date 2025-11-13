package tempvar

import (
	"errors"
	"testing"
	"time"

	"github.com/zalgonoise/x/is"
)

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time {
	return c.now
}

func TestTimeBased_Value(t *testing.T) {
	type user struct {
		name string
		id   int
		err  error
	}

	u := &user{
		name: "Gopher",
		id:   1,
	}

	clock := &testClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}

	for _, testcase := range []struct {
		name  string
		start time.Time
		end   time.Time
		wants *user
		err   error
	}{
		{
			name:  "Success/OnTime",
			start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC),
			wants: u,
		},
		{
			name:  "Success/BeforeStart",
			start: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			name:  "Success/AfterEnd",
			start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			name:  "Success/SameStartAndEndAdds1Second",
			start: time.Date(2025, 1, 1, 11, 59, 59, 1, time.UTC),
			end:   time.Date(2025, 1, 1, 11, 59, 59, 1, time.UTC),
			wants: u,
		},
		{
			name: "Fail/ZeroStart",
			end:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			err:  ErrZeroStart,
		},
		{
			name:  "Fail/ZeroEnd",
			start: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			err:   ErrZeroEnd,
		},
		{
			name:  "Fail/StartAfterEnd",
			start: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			err:   ErrStartAfterEnd,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tb, err := NewTimeBased(clock, testcase.start, testcase.end, *u)
			if err != nil {
				is.True(t, errors.Is(err, testcase.err))

				return
			}

			is.NilError(t, err)
			v := tb.Value()
			is.EqualValue(t, testcase.wants, v)
		})
	}
}

func TestTimeBased_Value_WithRealClock(t *testing.T) {
	u := &struct {
		name string
		id   int
	}{
		name: "Gopher",
		id:   1,
	}

	tb, err := NewTimeBased(nil,
		time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
		*u)

	is.NilError(t, err)
	v := tb.Value()
	is.EmptyValue(t, v)
}
