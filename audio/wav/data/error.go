package data

import "github.com/zalgonoise/x/audio/errs"

const (
	ErrDomain = errs.Domain("audio/wav/data")

	ErrInvalid = errs.Kind("invalid")
	ErrMissing = errs.Kind("missing")

	ErrBitDepth = errs.Entity("bit depth")
	ErrHeader   = errs.Entity("header")
)

var (
	ErrInvalidBitDepth = errs.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrMissingHeader   = errs.New(ErrDomain, ErrMissing, ErrHeader)
)
