package tempvar

import (
	"context"
	"sync/atomic"
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

type atom[T any] struct {
	v *atomic.Pointer[T]
}

func (a *atom[T]) Value() *T {
	return a.v.Load()
}

func newAtom[T any](ctx context.Context, value *T, dur time.Duration) *atom[T] {
	v := &atom[T]{v: &atomic.Pointer[T]{}}
	v.v.Store(value)

	go func() {
		t := time.NewTimer(dur)

		select {
		case <-t.C:
			v.v.Store(nil)
		case <-ctx.Done():
			v.v.Store(nil)
		}

		return
	}()

	return v
}

func BenchmarkTimed_Value(b *testing.B) {
	type user struct {
		name string
		id   int
	}

	for _, testcase := range []struct {
		name    string
		data    user
		dur     time.Duration
		newFunc func() interface{ Value() *user }
	}{
		{
			name: "Current",
			newFunc: func() interface{ Value() *user } {
				return NewTimedVar(context.Background(), &user{
					name: "Gopher",
					id:   1,
				}, time.Minute)
			},
		},
		{
			name: "Atomic",
			newFunc: func() interface{ Value() *user } {
				return newAtom(context.Background(), &user{
					name: "Gopher",
					id:   1,
				}, time.Minute)
			},
		},
	} {
		b.Run(testcase.name, func(b *testing.B) {
			v := testcase.newFunc()

			var value *user
			for b.Loop() {
				value = v.Value()
				if value == nil {
					return
				}
			}
		})
	}
}
