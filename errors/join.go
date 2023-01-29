package errors

import "strings"

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if errs contains no non-nil values.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
func Join(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &joinError{
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		if jerr, ok := err.(*joinError); ok {
			e.errs = append(e.errs, jerr.errs...)
			continue
		}
		e.errs = append(e.errs, err)
	}
	return e
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	var b = new(strings.Builder)
	for i, err := range e.errs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (e *joinError) Unwrap() []error {
	return e.errs
}
