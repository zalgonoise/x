package validation

import "errors"

// Validator is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unset data in the input data structure.
type Validator[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(T) error
}

// checker is a type of Validator that validates a data structure from a single or multiple validation functions
type checker[T any] []func(T) error

// Validate implements the Validator interface
//
// It executes the underlying functions in the Group v on the input data of type T, returning the errors that it finds.
//
// If the checker doesn't find any errors, it will return nil. If it finds one error, it will return the first item of
// an error slice it creates internally. If the call raises multiple errors, they are all collected in such slice,
// joined together (with errors.Join) and then returned to the caller as a single error.
func (v checker[T]) Validate(data T) error {
	if len(v) == 1 {
		return v[0](data)
	}

	errs := make([]error, 0, len(v))

	for i := range v {
		if err := v[i](data); err != nil {
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

// Register returns the input validation function or set of validation functions as a Validator of type T.
func Register[T any](validators ...func(T) error) Validator[T] {
	if len(validators) == 0 {
		return nil
	}

	return checker[T](validators)
}
