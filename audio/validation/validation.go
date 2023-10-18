package validation

import (
	"errors"
)

// Validator is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unset data in the input data structure.
type Validator[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(value T) error
}

type Func[T any] func(T) error

func (fn Func[T]) Validate(value T) error {
	return fn(value)
}

type multiValidator[T any] struct {
	validators []Validator[T]
}

func (v multiValidator[T]) Validate(value T) error {
	errs := make([]error, 0, len(v.validators))

	for i := range v.validators {
		if err := v.validators[i].Validate(value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func New[T any](validators ...func(T) error) Validator[T] {
	switch len(validators) {
	case 0:
		return noOpValidator[T]{}
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

func NoOp[T any]() Validator[T] {
	return noOpValidator[T]{}
}

type noOpValidator[T any] struct{}

func (noOpValidator[T]) Validate(T) error { return nil }
