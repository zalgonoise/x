package validation

// Validator is a generic type that matches an input item of comparable type T with a preselected list of
// T values that are considered valid.
//
// If there are zero matches, then its Validate function returns a (configurable) error. It returns nil on the
// first match encountered
type Validator[T comparable] struct {
	Values []T
	err    error
}

// Validate matches the input item against the slice of T configured in this Validator, returning nil on the first
// match, or the configured error if there are no matches
func (v *Validator[T]) Validate(item T) error {
	for i := range v.Values {
		if v.Values[i] == item {
			return nil
		}
	}

	return v.err
}

// New creates a Validator of type T, based on the given error and slice of values.
//
// While the function signature is variadic, it is mandatory that the slice of T `values` is not empty; as well as the
// provided error which should not be nil
func New[T comparable](err error, values ...T) *Validator[T] {
	if err == nil || len(values) == 0 {
		return nil
	}

	return &Validator[T]{
		Values: values,
		err:    err,
	}
}

// DynamicValidator is a generic type that matches an input item of any type T with a preselected list of
// T values that are considered valid. The comparison is done by a function that compares two T, returning a boolean
// if they are a match.
//
// If there are zero matches, then its Validate function returns a (configurable) error. It returns nil on the
// first match encountered
type DynamicValidator[T any] struct {
	Values []T
	fn     func(T, T) bool
	err    error
}

// Validate matches the input item against the slice of T configured in this DynamicValidator,
// using the configured function `func(T, T) bool` to compare each configured value to the input,
// returning nil on the first match, or the configured error if there are no matches
func (v *DynamicValidator[T]) Validate(item T) error {
	for i := range v.Values {
		if v.fn(item, v.Values[i]) {
			return nil
		}
	}

	return v.err
}

// With creates a DynamicValidator of type T, based on the given error, comparison function and slice of values.
//
// While the function signature is variadic, it is mandatory that the slice of T `values` is not empty; as well as the
// provided function and error which should not be nil
func With[T any](err error, fn func(input, target T) bool, values ...T) *DynamicValidator[T] {
	if err == nil || fn == nil || len(values) == 0 {
		return nil
	}

	return &DynamicValidator[T]{
		Values: values,
		fn:     fn,
		err:    err,
	}
}
