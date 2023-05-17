package wav

import (
	"github.com/zalgonoise/x/audio/errs"
)

const (
	ErrDomain = errs.Domain("audio/wav")

	ErrEmpty   = errs.Kind("missing")
	ErrInvalid = errs.Kind("invalid")
	ErrShort   = errs.Kind("short")

	ErrBitDepth   = errs.Entity("bit depth")
	ErrHeader     = errs.Entity("WAV header")
	ErrDataBuffer = errs.Entity("data buffer")
)

var (
	ErrInvalidBitDepth   = errs.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader     = errs.New(ErrDomain, ErrInvalid, ErrHeader)
	ErrShortDataBuffer   = errs.New(ErrDomain, ErrShort, ErrDataBuffer)
	ErrMissingHeader     = errs.New(ErrDomain, ErrEmpty, ErrHeader)
	ErrMissingDataBuffer = errs.New(ErrDomain, ErrInvalid, ErrDataBuffer)
)
