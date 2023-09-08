package errs

import (
	"errors"
	"fmt"
	"strings"
)

const (
	defaultFormat           = "%s %s"
	wrappedFormat           = "%s %s: %s"
	defaultFormatWithDomain = "%s: %s %s"
	wrappedFormatWithDomain = "%s: %s %s: %s"
)

type Domain string

func (e Domain) Error() string { return (string)(e) }

type Kind string

func (e Kind) Error() string { return (string)(e) }

type Entity string

func (e Entity) Error() string { return (string)(e) }

type Error struct {
	kind   Kind
	entity Entity

	format string
	inner  error
}

func (e Error) Error() string {
	format := e.format

	if e.inner == nil {
		if format == "" {
			format = defaultFormat
		}

		return fmt.Sprintf(format, e.kind, e.entity)
	}

	if format == "" {
		format = wrappedFormat
	}

	return fmt.Sprintf(format, e.kind, e.entity, e.inner.Error())
}

func (e Error) Unwrap() error {
	return e.inner
}

type ErrorWithDomain struct {
	err Error

	domain Domain
}

func (e ErrorWithDomain) Error() string {
	format := e.err.format

	if e.err.inner == nil {
		if format == "" {
			format = defaultFormatWithDomain
		}

		return fmt.Sprintf(format, e.domain, e.err.kind, e.err.entity)
	}

	if format == "" {
		format = wrappedFormatWithDomain
	}

	return fmt.Sprintf(format, e.domain, e.err.kind, e.err.entity, e.err.inner.Error())
}

func (e ErrorWithDomain) Unwrap() error {
	return e.err.inner
}

func New(domain Domain, kind Kind, entity Entity, args ...any) error {
	switch {
	case kind == "" && entity == "":
		return nil
	case kind == "" && entity != "":
		return entity
	case kind != "" && entity == "":
		return kind
	}

	return WithDomain(domain, Error{
		kind:   kind,
		entity: entity,
		inner:  joinArgs(args...),
	})
}

func Format(domain Domain, kind Kind, entity Entity, format string, args ...any) error {
	err := New(domain, kind, entity, args...)

	if format == "" {
		return err
	}

	switch e := err.(type) {
	case Error:
		e.format = format

		return e
	case ErrorWithDomain:
		e.err.format = format

		return e
	default:
		return err
	}
}

func WithDomain(domain Domain, err Error) error {
	if domain == "" {
		return err
	}

	return ErrorWithDomain{
		err:    err,
		domain: domain,
	}
}

func joinArgs(args ...any) error {
	if len(args) == 0 {
		return nil
	}

	format := &strings.Builder{}
	errs := make([]any, 0, len(args))

	for i := range args {
		switch v := args[i].(type) {
		case nil:
			continue

		case ErrorWithDomain:
			format.WriteString("%w")
			if i < len(args)-1 {
				format.WriteString(": ")
			}

			errs = append(errs, errors.New(fmt.Sprintf(defaultFormatWithDomain, v.domain, v.err.kind, v.err.entity)))

		case error:
			format.WriteString("%w")
			if i < len(args)-1 {
				format.WriteString(": ")
			}

			errs = append(errs, v)

		default:
			format.WriteString("%w")
			if i < len(args)-1 {
				format.WriteString(": ")
			}

			errs = append(errs, errors.New(fmt.Sprint(v)))
		}
	}

	return fmt.Errorf(format.String(), errs...)
}
