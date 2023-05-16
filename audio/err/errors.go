package err

import (
	"errors"
	"fmt"
	"strings"
)

const defaultDomain = "audio/wav"

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

func New(domain Domain, kind Kind, entity Entity) error {
	if domain == "" {
		domain = defaultDomain
	}

	sb := new(strings.Builder)
	sb.WriteString((string)(domain))
	sb.WriteString(": ")
	sb.WriteString((string)(kind))
	sb.WriteByte(' ')
	sb.WriteString((string)(entity))

	return Error{
		kind:   kind,
		entity: entity,
		error:  sb.String(),
	}
}

func Errorf(domain Domain, kind Kind, entity Entity, args ...any) error {
	var err string

	if domain == "" {
		domain = defaultDomain
	}

	sb := new(strings.Builder)
	sb.WriteString((string)(domain))
	sb.WriteString(": ")
	sb.WriteString((string)(kind))
	sb.WriteByte(' ')
	sb.WriteString((string)(entity))
	err = sb.String()

	if len(args) > 0 {
		err = fmt.Sprint(err, args)
	}

	return Error{
		kind:   kind,
		entity: entity,
		error:  err,
	}
}
