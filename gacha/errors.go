package gacha

import "github.com/zalgonoise/x/gacha/internal/errs"

// Sentinel error kinds and entities used across gacha.
const (
	libDomain    = "gacha"
	statusDomain = "gacha/Status"

	ErrInvalid = errs.Kind("invalid")

	ErrStatus = errs.Entity("status")
	ErrLease  = errs.Entity("lease")
)

// Sentinel errors returned by gacha components.
var (
	// ErrInvalidStatus is returned when parsing an unknown status, and by Source.Settle when an outcome carries a
	// status other than StatusCompleted, StatusPending or StatusFailed.
	ErrInvalidStatus = errs.New(statusDomain, ErrInvalid, ErrStatus)

	// ErrInvalidLease is returned by Source.Claim and Extender.Extend when the lease duration is not positive.
	ErrInvalidLease = errs.New(libDomain, ErrInvalid, ErrLease)
)
