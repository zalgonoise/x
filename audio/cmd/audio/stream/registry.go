package stream

import (
	"maps"
)

// LessFunc is a generic, injectable function that determines how to
// compare two values of any type, returning a boolean on whether
// item `i` is less than item `j`
type LessFunc[T any] func(i, j T) bool

type Registry[T any, F any] interface {
	Register(T)
	Flush() F
}

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
	maximum := r.max
	r.max = *new(T)

	return maximum
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

type LabeledRegistry[T any, F ~map[string]T] struct {
	max   map[string]T
	less  LessFunc[T]
	label func(T) string
}

func (r LabeledRegistry[T, F]) Register(value T) {
	label := r.label(value)
	maximum, ok := r.max[label]
	if ok && r.less(maximum, value) {
		r.max[label] = value

		return
	}

	r.max[label] = value
}

func (r LabeledRegistry[T, F]) Flush() F {
	values := make(map[string]T, len(r.max))
	maps.Copy(values, r.max)

	for k := range r.max {
		r.max[k] = *new(T)
	}

	return values
}

func NewLabeledRegistry[T any, F ~map[string]T](lessFunc LessFunc[T], labelFunc func(T) string) LabeledRegistry[T, F] {
	return LabeledRegistry[T, F]{
		max:   make(map[string]T, 2048),
		less:  lessFunc,
		label: labelFunc,
	}
}
