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

func NewExhaustedVar[T any](value *T, limit uint64) *Exhausted[T] {
	if limit < minLimit {
		limit = minLimit
	}

	count := &atomic.Uint64{}

	return &Exhausted[T]{
		value: value,
		limit: limit,
		count: count,
	}
}

func (v *Exhausted[T]) Value() *T {
	v.count.Add(1)

	if v.count.Load() <= v.limit {
		return v.value
	}

	return nil
}
