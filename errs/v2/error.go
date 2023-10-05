package errs

import "errors"

var (
	spaceSeparator = []rune{' '}
	colonSeparator = []rune{':', ' '}
)

type errsType interface {
	error
	Unwrap() error
}

type Domain string

func (e Domain) Error() string { return (string)(e) }
func (e Domain) Unwrap() error { return nil }

type Kind string

func (e Kind) Error() string { return (string)(e) }
func (e Kind) Unwrap() error { return nil }

type Entity string

func (e Entity) Error() string { return (string)(e) }
func (e Entity) Unwrap() error { return nil }

func newSentinel(domain Domain, kind Kind, entity Entity) error {
	switch {
	case kind == "" && entity == "":
		return nil
	case domain == "" && kind == "":
		return entity
	case domain == "" && entity == "":
		return kind
	}

	s := sentinel{
		kind:   kind,
		entity: entity,

		errString: sentinelErrString(kind, entity),
	}

	if domain == "" {
		return s
	}

	return sentinelWithDomain{
		sentinel:  s,
		domain:    domain,
		errString: sentinelWithDomainErrString(domain, s.errString),
	}
}

func sentinelErrString(kind Kind, entity Entity) string {
	s := make([]rune, len(kind)+len(spaceSeparator)+len(entity))

	n := copy(s, []rune(kind))
	n += copy(s[n:], spaceSeparator)
	copy(s[n:], []rune(entity))

	return string(s)
}

func sentinelWithDomainErrString(domain Domain, err string) string {
	s := make([]rune, len(domain)+len(colonSeparator)+len(err))

	n := copy(s, []rune(domain))
	n += copy(s[n:], colonSeparator)
	n += copy(s, []rune(err))

	return string(s)
}

type sentinel struct {
	kind   Kind
	entity Entity

	errString string
}

func (e sentinel) Error() string {
	return e.errString
}

func (e sentinel) Unwrap() error {
	return errors.Join(e.kind, e.entity)
}

type sentinelWithDomain struct {
	sentinel

	domain    Domain
	errString string
}

func (e sentinelWithDomain) Error() string {
	return e.errString
}

func (e sentinelWithDomain) Unwrap() error {
	return errors.Join(e.domain, e.kind, e.entity)
}

type sentinelWrapped struct {
	err errsType

	wrapped   error
	errString string
}

func (e sentinelWrapped) Error() string {
	return e.errString
}

func (e sentinelWrapped) Unwrap() error {
	return errors.Join(e.err.Unwrap(), e.wrapped)
}
