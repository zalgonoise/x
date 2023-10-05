package errs

import "fmt"

var (
	spaceSeparator = []rune{' '}
	colonSeparator = []rune{':', ' '}
)

type errsType interface {
	// Error allows a simple (and standard) approach to creating
	// an error type.
	error
	// Unwrap allows a multi-error extraction method, returning a set of all
	// errors that compose this type.
	//
	// It can safely be used in an errors.Is call.
	Unwrap() []error
}

type Domain string

func (e Domain) Error() string   { return (string)(e) }
func (e Domain) Unwrap() []error { return nil }

type Kind string

func (e Kind) Error() string   { return (string)(e) }
func (e Kind) Unwrap() []error { return nil }

type Entity string

func (e Entity) Error() string   { return (string)(e) }
func (e Entity) Unwrap() []error { return nil }

func Errorf(format string, args ...any) error {
	if len(args) == 0 {
		return fmt.Errorf(format)
	}

	// strip domains in Errorf
	for i := range args {
		if a, ok := (args[i]).(sentinelWithDomain); ok {
			args[i] = a.sentinel
		}
	}

	return fmt.Errorf(format, args...)
}

func Sentinel(kind Kind, entity Entity) error {
	return newSentinel("", kind, entity)
}

func WithDomain(domain Domain, kind Kind, entity Entity) error {
	return newSentinel(domain, kind, entity)
}

func Wrap(err errsType, wrapper error) error {
	return wrapErrorWithoutDomain(err, wrapper)
}

func withSpace[K, E ~string](first K, last E) string {
	s := make([]rune, len(first)+len(spaceSeparator)+len(last))

	n := copy(s, []rune(first))
	n += copy(s[n:], spaceSeparator)
	copy(s[n:], []rune(last))

	return string(s)
}

func withColon[K, E ~string](first K, last E) string {
	s := make([]rune, len(first)+len(colonSeparator)+len(last))

	n := copy(s, []rune(first))
	n += copy(s[n:], colonSeparator)
	n += copy(s, []rune(last))

	return string(s)
}
