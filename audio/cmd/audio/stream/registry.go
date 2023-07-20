package stream

type MaxRegistry[T any] struct {
	max T

	lessFn func(i, j T) bool
}

func (r *MaxRegistry[T]) Register(value T) {
	if r.lessFn(r.max, value) {
		r.max = value
	}
}

func (r *MaxRegistry[T]) Flush() T {
	max := r.max
	r.max = *new(T)

	return max
}

func NewMaxRegistry[T any](lessFn func(i, j T) bool) *MaxRegistry[T] {
	if lessFn == nil {
		return nil
	}

	return &MaxRegistry[T]{
		max:    *new(T),
		lessFn: lessFn,
	}
}
