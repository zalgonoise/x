// Package errs provides structured sentinel errors decomposed into Domain, Kind, and Entity.
//
// It mirrors github.com/zalgonoise/gbuf/v2/errs, with "gacha" as the default domain.
package errs

import (
	"errors"
	"strings"
)

const defaultDomain = "gacha"

// Domain identifies the package or component where an error originated.
type Domain string

func (e Domain) Error() string { return (string)(e) }

// Kind describes the category of error that occurred.
type Kind string

func (e Kind) Error() string { return (string)(e) }

// Entity identifies the target value or operation associated with an error.
type Entity string

func (e Entity) Error() string { return (string)(e) }

// Error is a structured sentinel error composed of a Domain, Kind, and Entity.
type Error struct {
	domain Domain
	kind   Kind
	entity Entity
	error  string
}

func (e Error) Error() string {
	return e.error
}

// Is reports whether target matches the Error or any of its Domain, Kind, or Entity components.
func (e Error) Is(target error) bool {
	return target == e || target == e.domain || target == e.kind || target == e.entity
}

// Unwrap returns the joined Domain, Kind, and Entity errors.
func (e Error) Unwrap() error {
	return errors.Join(e.domain, e.kind, e.entity)
}

// New constructs a structured Error from the given domain, kind, and entity.
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
		domain: domain,
		kind:   kind,
		entity: entity,
		error:  sb.String(),
	}
}
