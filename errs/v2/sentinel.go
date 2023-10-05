package errs

type sentinel struct {
	kind   Kind
	entity Entity

	errString string
}

func (e sentinel) Error() string {
	return e.errString
}

func (e sentinel) Unwrap() []error {
	return []error{e.kind, e.entity}
}

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

		errString: withSpace(kind, entity),
	}

	if domain == "" {
		return s
	}

	return sentinelWithDomain{
		sentinel:  s,
		domain:    domain,
		errString: withColon(domain, s.errString),
	}
}
