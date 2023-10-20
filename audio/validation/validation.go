package validation

import (
	"errors"
)

// Validator is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unexpectedly unset data in the input data structure.
type Validator[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(value T) error
}

// Func is a function type that complies with the Validator's Validate method signature.
//
// The Func type implements the Validator interface, through a Validate method calling on itself.
type Func[T any] func(T) error

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner Func with the input value.
func (fn Func[T]) Validate(value T) error {
	return fn(value)
}

// Funcs is a joined function type that complies with the Validator's Validate method signature.
//
// The Funcs type implements the Validator interface, through a Validate method calling on all the functions in the
// slice.
type Funcs[T any] []func(T) error

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner functions in the Funcs type, with the input value.
func (fns Funcs[T]) Validate(value T) error {
	errs := make([]error, 0, len(fns))

	for i := range fns {
		if fns[i] == nil {
			continue
		}

		if err := fns[i](value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type multiValidator[T any] struct {
	validators []Validator[T]
}

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by iterating
// through all configured Validator, while calling their Validate method on the input value.
func (v multiValidator[T]) Validate(value T) error {
	errs := make([]error, 0, len(v.validators))

	for i := range v.validators {
		if err := v.validators[i].Validate(value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// NewValidator creates a Validator from the input slice of Func.
//
// If the input slice contains no items, this call returns a NoOp Validator. If it only contains one function, it
// will return it as a Func type, effectively as a Validator.
//
// If there are multiple functions in the input, a multi-Validator is created. This multi-Validator will contain all
// non-nil validators from the input, that will work with the same input value in one go.
func NewValidator[T any](validators ...func(T) error) Validator[T] {
	switch len(validators) {
	case 0:
		return NoOp[T]()
	case 1:
		return Func[T](validators[0])
	}

	c := make([]Validator[T], 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		c = append(c, Func[T](validators[i]))

	}

	return &multiValidator[T]{
		validators: c,
	}
}

// Join gathers multiple Validator for the same type, joining them as a single Validator. It is similar to NewValidator,
// but works exclusively with Validator types as input.
func Join[T any](validators ...Validator[T]) Validator[T] {
	switch len(validators) {
	case 0:
		return NoOp[T]()
	case 1:
		return validators[0]
	}

	c := make([]Validator[T], 0, len(validators))

	for i := range validators {
		switch v := validators[i].(type) {
		case nil:
			continue
		case multiValidator[T]:
			c = append(c, v.validators...)
		default:
			c = append(c, v)
		}
	}

	return &multiValidator[T]{
		validators: c,
	}
}

// NoOp returns a no-op Validator.
func NoOp[T any]() Validator[T] {
	return noOpValidator[T]{}
}

type noOpValidator[T any] struct{}

// Validate implements the Validator interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpValidator[T]) Validate(T) error { return nil }
