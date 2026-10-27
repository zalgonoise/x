package gacha

import "errors"

type permanentError struct {
	err error
}

func (e permanentError) Error() string {
	return e.err.Error()
}

func (e permanentError) Unwrap() error {
	return e.err
}

func (permanentError) Permanent() bool {
	return true
}

// Permanent marks err as unrecoverable: when a handler returns it, the poller moves the entry to the dead-letter
// table immediately instead of retrying it. The returned error has the same message as err and wraps it.
//
// Permanent returns nil if err is nil.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err: err}
}

// IsPermanent reports whether err is unrecoverable: whether the first error in its tree with a Permanent() bool
// method reports true. Errors returned by Permanent always do, and other error types can implement the method to opt
// in.
func IsPermanent(err error) bool {
	var permanent interface{ Permanent() bool }

	return errors.As(err, &permanent) && permanent.Permanent()
}
