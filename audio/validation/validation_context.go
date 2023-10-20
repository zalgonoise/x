package validation

import (
	"context"
	"errors"
)

// ValidatorContext is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unexpectedly unset data in the input data structure
// using a context.Context in its Validate call.
type ValidatorContext[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(ctx context.Context, value T) error
}

// FuncContext is a function type that complies with the ValidatorContext's Validate method signature.
//
// The FuncContext type implements the ValidatorContext interface, through a Validate method calling on itself.
type FuncContext[T any] func(context.Context, T) error

// Validate implements the ValidatorContext interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner FuncContext with the input value.
func (fn FuncContext[T]) Validate(ctx context.Context, value T) error {
	return fn(ctx, value)
}

// FuncsContext is a joined function type that complies with the Validator's Validate method signature.
//
// The FuncsContext type implements the ValidatorContext interface, through a Validate method calling on all the
// functions in the slice.
type FuncsContext[T any] []func(context.Context, T) error

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner functions in the FuncsContext type, with the input value.
func (fns FuncsContext[T]) Validate(ctx context.Context, value T) error {
	errs := make([]error, 0, len(fns))

	for i := range fns {
		if fns[i] == nil {
			continue
		}

		if err := fns[i](ctx, value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type multiValidatorContext[T any] struct {
	validators []ValidatorContext[T]
}

// Validate implements the ValidatorContext interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by iterating
// through all configured ValidatorContext, while calling their Validate method on the input value.
func (v multiValidatorContext[T]) Validate(ctx context.Context, value T) error {
	errs := make([]error, 0, len(v.validators))

	for i := range v.validators {
		if err := v.validators[i].Validate(ctx, value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// NewValidatorContext creates a ValidatorContext from the input slice of FuncContext.
//
// If the input slice contains no items, this call returns a NoOpContext ValidatorContext. If it only contains one
// function, it will return it as a FuncContext type, effectively as a ValidatorContext.
//
// If there are multiple functions in the input, a multi-Validator is created. This multi-Validator will contain all
// non-nil validators from the input, that will work with the same input value in one go.
func NewValidatorContext[T any](validators ...func(context.Context, T) error) ValidatorContext[T] {
	switch len(validators) {
	case 0:
		return NoOpContext[T]()
	case 1:
		return FuncContext[T](validators[0])
	}

	c := make([]ValidatorContext[T], 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		c = append(c, FuncContext[T](validators[i]))

	}

	return &multiValidatorContext[T]{
		validators: c,
	}
}

// JoinContext gathers multiple ValidatorContext for the same type, joining them as a single ValidatorContext.
// It is similar to NewValidatorContext, but works exclusively with ValidatorContext types as input.
func JoinContext[T any](validators ...ValidatorContext[T]) ValidatorContext[T] {
	switch len(validators) {
	case 0:
		return NoOpContext[T]()
	case 1:
		return validators[0]
	}

	c := make([]ValidatorContext[T], 0, len(validators))

	for i := range validators {
		switch v := validators[i].(type) {
		case nil:
			continue
		case multiValidatorContext[T]:
			c = append(c, v.validators...)
		default:
			c = append(c, v)
		}
	}

	return &multiValidatorContext[T]{
		validators: c,
	}
}

// NoOpContext returns a no-op ValidatorContext.
func NoOpContext[T any]() ValidatorContext[T] {
	return noOpValidatorContext[T]{}
}

type noOpValidatorContext[T any] struct{}

// Validate implements the ValidatorContext interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpValidatorContext[T]) Validate(context.Context, T) error { return nil }
