package data

import "github.com/zalgonoise/x/audio/errs"

const (
	ErrDomain = errs.Domain("audio/wav/data")

	ErrInvalid = errs.Kind("invalid")

	ErrBitDepth = errs.Entity("bit depth")
)

var ErrInvalidBitDepth = errs.New(ErrDomain, ErrInvalid, ErrBitDepth)
