package header

import (
	"github.com/zalgonoise/x/audio/errs"
)

const (
	ErrDomain = errs.Domain("audio/wav/header")

	ErrEmpty   = errs.Kind("missing")
	ErrInvalid = errs.Kind("invalid")
	ErrShort   = errs.Kind("short")

	ErrNumChannels = errs.Entity("number of channels")
	ErrSampleRate  = errs.Entity("sample rate")
	ErrBitDepth    = errs.Entity("bit depth")
	ErrHeader      = errs.Entity("WAV header")
	ErrDataBuffer  = errs.Entity("data buffer")
	ErrAudioFormat = errs.Entity("audio format")
)

var (
	ErrInvalidNumChannels = errs.New(ErrDomain, ErrInvalid, ErrNumChannels)
	ErrInvalidSampleRate  = errs.New(ErrDomain, ErrInvalid, ErrSampleRate)
	ErrInvalidBitDepth    = errs.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader      = errs.New(ErrDomain, ErrInvalid, ErrHeader)
	ErrShortDataBuffer    = errs.New(ErrDomain, ErrShort, ErrDataBuffer)
	ErrMissingHeader      = errs.New(ErrDomain, ErrEmpty, ErrHeader)
	ErrMissingDataBuffer  = errs.New(ErrDomain, ErrInvalid, ErrDataBuffer)
	ErrInvalidAudioFormat = errs.New(ErrDomain, ErrInvalid, ErrAudioFormat)
)
