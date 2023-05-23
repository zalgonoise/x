package validation

import "errors"

// ProcessorFunc is a generic function type implemented by the caller of New or FailFast, which analyzes an item of
// any type to validate it on some level; returning an error if invalid
type ProcessorFunc[T any] func(item T) error

// Validator is a generic data structure that handles multi-pass / multi-element validation of any given type, as
// implemented by the caller.
//
// The Validator iterates through its configured ProcessorFunc over a given item, reporting back to the caller on any
// validity errors raised
type Validator[T any] struct {
	failFast bool
	fn       []ProcessorFunc[T]
}

// Validate iterates through the configured ProcessorFunc to validate the input `item` of type T.
//
// This execution exits on the first error encountered if the Validator was created using FailFast; otherwise it
// will collect a list of errors found, joining them (with errors.Join) if more than one error is found.
func (v *Validator[T]) Validate(item T) error {
	errs := make([]error, 0, len(v.fn))

	for i := range v.fn {
		if err := v.fn[i](item); err != nil {
			if v.failFast {
				return err
			}

			errs = append(errs, err)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.Join(errs...)
	}
}

// New creates a plain Validator for type T, configured with the input ProcessorFunc.
//
// This call returns nil if no ProcessorFunc is provided.
func New[T any](fn ...ProcessorFunc[T]) *Validator[T] {
	if len(fn) == 0 {
		return nil
	}

	return &Validator[T]{
		fn: fn,
	}
}

// FailFast creates a Validator for type T, configured with the input ProcessorFunc,
// that exits on the first error raised instead of iterating through the entire set of ProcessorFunc.
//
// This call returns nil if no ProcessorFunc is provided.
func FailFast[T any](fn ...ProcessorFunc[T]) *Validator[T] {
	if len(fn) == 0 {
		return nil
	}

	return &Validator[T]{
		failFast: true,
		fn:       fn,
	}
}
