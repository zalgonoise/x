package tempvar

import (
	"sync/atomic"
)

const (
	minLimit uint64 = 1
)

type Exhausted[T any] struct {
	value *T
	limit uint64
	count *atomic.Uint64
}

func NewExhaustedVar[T any](value T, limit uint64) *Exhausted[T] {
	if limit < minLimit {
		limit = minLimit
	}

	return &Exhausted[T]{
		value: &value,
		limit: limit,
		count: &atomic.Uint64{},
	}
}

func (v *Exhausted[T]) Value() *T {
	if v.count.Load() < v.limit {
		v.count.Add(1)

		return v.value
	}

	return nil
}
