package errs

type wrapped struct {
	err errsType

	wrapper   error
	errString string
}

func (e wrapped) Error() string {
	return e.errString
}

func (e wrapped) Unwrap() []error {
	return append(e.err.Unwrap(), e.wrapper)
}

func wrapError(err errsType, wrapper error) error {
	return wrapped{
		err:       err,
		wrapper:   wrapper,
		errString: withColon(wrapper.Error(), err.Error()),
	}
}

func wrapErrorWithoutDomain(err errsType, wrapper error) error {
	if e, ok := err.(sentinelWithDomain); ok {
		err = e.sentinel
	}

	return wrapped{
		err:       err,
		wrapper:   wrapper,
		errString: withColon(wrapper.Error(), err.Error()),
	}
}
