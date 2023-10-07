package errs

import (
	"errors"
	"strings"
)

const (
	defaultSeparator = ": "
)

func Join(e ...error) error {
	return join(defaultSeparator, e...)
}

func JoinWith(sep string, e ...error) error {
	if sep == "" {
		sep = defaultSeparator
	}

	return join(sep, e...)
}

func compact(e ...error) []error {
	n := 0
	for _, err := range e {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}

	output := make([]error, 0, n)

	for i := range e {
		if e[i] != nil {
			output = append(output, e[i])
		}
	}
	return output
}

func join(sep string, e ...error) error {
	e = compact(e...)

	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	}

	sb := &strings.Builder{}

	for i := range e {
		printError(sb, e[i])

		if i < len(e)-1 {
			sb.WriteString(sep)
		}
	}

	return joinedError{
		errString: sb.String(),
		e:         e,
	}
}

func printError(sb *strings.Builder, err error) {
	var withDomain sentinelWithDomain
	if errors.As(err, &withDomain) {
		sb.WriteString(withDomain.sentinel.errString)

		return
	}

	sb.WriteString(err.Error())
}

type joinedError struct {
	errString string
	e         []error
}

func (e joinedError) Error() string {
	return e.errString
}

func (e joinedError) Unwrap() []error {
	return e.e
}
