package stream

import (
	"errors"

	"golang.org/x/exp/maps"
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

type BucketRegistry[T any, F ~map[string]T] struct {
	max map[string]T

	bucket bucket[T]
	less   LessFunc[T]
}

func (r BucketRegistry[T, F]) Register(value T) {
	if label := r.bucket.Get(value); label != "" {
		if max, ok := r.max[label]; ok && r.less(max, value) {
			r.max[label] = value
		}
	}
}

func (r BucketRegistry[T, F]) Flush() F {
	values := make(map[string]T, len(r.max))
	maps.Copy(values, r.max)

	for k := range r.max {
		r.max[k] = *new(T)
	}

	return values
}

func NewBucketRegistry[T any, F ~map[string]T](
	values []T, labels []string, lessFunc LessFunc[T],
) (r BucketRegistry[T, F], err error) {
	switch {
	case len(labels) == 0 && len(values) == 0, len(labels) > 0 && len(values) == 0:
		//labels = frequencyLabels
		//values = make([]T, len(frequencyValues))
		//for i := range values {
		//	values[i] = T(frequencyValues[i])
		//}
		return r, errors.New("invalid set of labels and values")

	case len(labels) == 0 && len(values) > 0:
		//labels = make([]string, len(values))
		//for i := range labels {
		//	labels[i] = fmt.Sprintf("%d", int(values[i]))
		//}
		return r, errors.New("invalid set of labels and values")
	}

	max := make(map[string]T, len(labels))
	for i := range labels {
		max[labels[i]] = *new(T)
	}

	b := newBucket[T](values, labels, lessFunc)
	if b == nil {
		return r, errors.New("failed to create bucket")
	}

	return BucketRegistry[T, F]{
		max:    max,
		bucket: *b,
		less:   lessFunc,
	}, nil
}
