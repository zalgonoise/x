package stream

// LessFunc is a generic, injectable function that determines how to
// compare two values of any type, returning a boolean on whether
// item `i` is less than item `j`
type LessFunc[T any] func(i, j T) bool

// MaxRegistry stores the maximum value for a series of writes, as
// determined by the configured sort.LessFunc provided
type MaxRegistry[T any] struct {
	max T

	lessFn LessFunc[T]
}

// Register stores the input value as the max value in the MaxRegistry,
// if it is in fact greater than the stored one
func (r *MaxRegistry[T]) Register(value T) {
	if r.lessFn(r.max, value) {
		r.max = value
	}
}

// Flush returns the max value in the MaxRegistry, while resetting the one
// it has stored
func (r *MaxRegistry[T]) Flush() T {
	max := r.max
	r.max = *new(T)

	return max
}

// NewMaxRegistry creates a MaxRegistry from the input LessFunc
func NewMaxRegistry[T any](lessFn LessFunc[T]) *MaxRegistry[T] {
	if lessFn == nil {
		return nil
	}

	return &MaxRegistry[T]{
		max:    *new(T),
		lessFn: lessFn,
	}
}
