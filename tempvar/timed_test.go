package tempvar

import (
	"context"
	"testing"
	"time"

	"github.com/zalgonoise/x/is"
)

func TestTimed_Value(t *testing.T) {
	type user struct {
		name string
		id   int
	}

	for _, testcase := range []struct {
		name string
		data user
		dur  time.Duration
	}{
		{
			name: "ValueAndExpiry",
			data: user{
				name: "Gopher",
				id:   1,
			},
			dur: minDuration,
		},
		{
			name: "InvalidDuration",
			data: user{
				name: "Go",
				id:   2,
			},
			dur: 1 * time.Millisecond,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			ctx := context.Background()

			v := NewTimedVar(ctx, &testcase.data, testcase.dur)
			first := v.Value()

			time.Sleep(minDuration)
			second := v.Value()

			is.Equal(t, *first, testcase.data)
			is.Empty(t, second)
		})
	}
}
