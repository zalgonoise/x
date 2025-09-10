package tempvar

import (
	"context"
	"sync/atomic"
	"time"
)

const minDuration = 10 * time.Millisecond

type Timed[T any] struct {
	value     *T
	isExpired *atomic.Bool
}

func NewTimedVar[T any](ctx context.Context, value *T, dur time.Duration) *Timed[T] {
	v := &Timed[T]{
		value:     value,
		isExpired: &atomic.Bool{},
	}

	if dur < minDuration {
		dur = minDuration
	}

	go expireTimedVar(ctx, v, dur)

	return v
}

func (v *Timed[T]) Value() *T {
	if !v.isExpired.Load() {
		return v.value
	}

	return nil
}

func expireTimedVar[T any](ctx context.Context, v *Timed[T], dur time.Duration) {
	ctx, done := context.WithTimeout(ctx, dur)
	defer done()

	<-ctx.Done()
	v.isExpired.Store(true)
}
