package tempvar

import (
	"errors"
	"sync/atomic"
	"time"
)

const minTimeBasedDuration = time.Second

var (
	ErrZeroStart     = errors.New("start is zero")
	ErrZeroEnd       = errors.New("end is zero")
	ErrStartAfterEnd = errors.New("start is after end")
)

type Clock interface {
	Now() time.Time
}

type TimeBased[T any] struct {
	clock      Clock
	start, end time.Time
	value      *T
}

func NewTimeBased[T any](clock Clock, start, end time.Time, value T) (*TimeBased[T], error) {
	if start.Equal(time.Time{}) || start.IsZero() {
		return nil, ErrZeroStart
	}

	if end.Equal(time.Time{}) || end.IsZero() {
		return nil, ErrZeroEnd
	}

	if start.After(end) {
		return nil, ErrStartAfterEnd
	}

	if end.Sub(start) < minTimeBasedDuration {
		end = start.Add(minTimeBasedDuration)
	}

	if clock == nil {
		clock = RealClock{}
	}

	v := &atomic.Pointer[T]{}
	v.Store(&value)

	return &TimeBased[T]{
		clock: clock,
		start: start,
		end:   end,
		value: &value,
	}, nil
}

func (v *TimeBased[T]) Value() *T {
	now := v.clock.Now()

	if now.Before(v.start) || now.After(v.end) {
		return nil
	}

	return v.value
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}
