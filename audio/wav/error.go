package wav

import (
	"github.com/zalgonoise/x/audio/err"
)

const (
	ErrDomain = err.Domain("audio/wav")

	ErrEmpty   = err.Kind("missing")
	ErrInvalid = err.Kind("invalid")
	ErrShort   = err.Kind("short")

	ErrBitDepth   = err.Entity("bit depth")
	ErrHeader     = err.Entity("WAV header")
	ErrDataBuffer = err.Entity("data buffer")
)

var (
	ErrInvalidBitDepth   = err.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader     = err.New(ErrDomain, ErrInvalid, ErrHeader)
	ErrShortDataBuffer   = err.New(ErrDomain, ErrShort, ErrDataBuffer)
	ErrMissingHeader     = err.New(ErrDomain, ErrEmpty, ErrHeader)
	ErrMissingDataBuffer = err.New(ErrDomain, ErrInvalid, ErrDataBuffer)
)
