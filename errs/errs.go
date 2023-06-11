package errs

import (
	"errors"
	"fmt"
	"strings"
)

type Domain string

func (e Domain) Error() string { return (string)(e) }

type Kind string

func (e Kind) Error() string { return (string)(e) }

type Entity string

func (e Entity) Error() string { return (string)(e) }

type Error struct {
	domain Domain
	kind   Kind
	entity Entity
	error  string
}

func (e Error) Error() string {
	return e.error
}

func (e Error) Unwrap() error {
	return errors.Join(e.domain, e.kind, e.entity)
}

func New(domain Domain, kind Kind, entity Entity, args ...any) error {
	sb := new(strings.Builder)

	if domain != "" {
		sb.WriteString((string)(domain))
		sb.WriteString(": ")
	}

	sb.WriteString((string)(kind))
	sb.WriteByte(' ')
	sb.WriteString((string)(entity))

	if len(args) > 0 {
		_, _ = fmt.Fprint(sb, args)
	}

	return Error{
		domain: domain,
		kind:   kind,
		entity: entity,
		error:  sb.String(),
	}
}
