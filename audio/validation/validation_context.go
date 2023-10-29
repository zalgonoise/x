package validation

import (
	"context"
	"errors"
)

// ContextValidator is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unexpectedly unset data in the input data structure
// using a context.Context in its Validate call.
type ContextValidator[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(ctx context.Context, value T) error
}

// FuncContext is a function type that complies with the ContextValidator's Validate method signature.
//
// The FuncContext type implements the ContextValidator interface, through a Validate method calling on itself.
type FuncContext[T any] func(context.Context, T) error

// Validate implements the ContextValidator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner FuncContext with the input value.
func (fn FuncContext[T]) Validate(ctx context.Context, value T) error {
	if fn == nil {
		return nil
	}

	return fn(ctx, value)
}

type multiContextValidator[T any] struct {
	validators []ContextValidator[T]
}

// Validate implements the ContextValidator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by iterating
// through all configured ContextValidator, while calling their Validate method on the input value.
func (v multiContextValidator[T]) Validate(ctx context.Context, value T) error {
	errs := make([]error, 0, len(v.validators))

	for i := range v.validators {
		if err := v.validators[i].Validate(ctx, value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// NewContext creates a ContextValidator from the input slice of FuncContext.
//
// If the input slice contains no items, this call returns a NoOpContext ContextValidator. If it only contains one
// function, it will return it as a FuncContext type, effectively as a ContextValidator.
//
// If there are multiple functions in the input, a multi-Validator is created. This multi-Validator will contain all
// non-nil validators from the input, that will work with the same input value in one go.
func NewContext[T any](validators ...func(context.Context, T) error) ContextValidator[T] {
	validators = nonNilFuncContext(validators)

	switch len(validators) {
	case 0:
		return NoOpContext[T]()
	case 1:
		return FuncContext[T](validators[0])
	}

	mv := multiContextValidator[T]{
		validators: make([]ContextValidator[T], 0, len(validators)),
	}

	for i := range validators {
		mv.validators = append(mv.validators, FuncContext[T](validators[i]))
	}

	return mv
}

// JoinContext gathers multiple ContextValidator for the same type, joining them as a single ContextValidator.
// It is similar to NewContext, but works exclusively with ContextValidator types as input.
func JoinContext[T any](validators ...ContextValidator[T]) ContextValidator[T] {
	validators = nonNilContextValidator(validators)

	switch len(validators) {
	case 0:
		return NoOpContext[T]()
	case 1:
		return validators[0]
	}

	mv := multiContextValidator[T]{
		validators: make([]ContextValidator[T], 0, len(validators)),
	}

	for i := range validators {
		switch v := validators[i].(type) {
		case nil:
			continue
		case multiContextValidator[T]:
			mv.validators = append(mv.validators, v.validators...)
		default:
			mv.validators = append(mv.validators, v)
		}
	}

	return mv
}

// NoOpContext returns a no-op ContextValidator.
func NoOpContext[T any]() ContextValidator[T] {
	return noOpContextValidator[T]{}
}

type noOpContextValidator[T any] struct{}

// Validate implements the ContextValidator interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpContextValidator[T]) Validate(context.Context, T) error { return nil }

func nonNilFuncContext[T any](validators []func(context.Context, T) error) []func(context.Context, T) error {
	squash := make([]func(context.Context, T) error, 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		squash = append(squash, validators[i])
	}

	return squash
}

func nonNilContextValidator[T any](validators []ContextValidator[T]) []ContextValidator[T] {
	squash := make([]ContextValidator[T], 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		squash = append(squash, validators[i])
	}

	return squash
}
