package errs

type sentinelWithDomain struct {
	sentinel

	domain    Domain
	errString string
}

func (e sentinelWithDomain) Error() string {
	return e.errString
}

func (e sentinelWithDomain) Unwrap() []error {
	return []error{e.domain, e.kind, e.entity}
}
