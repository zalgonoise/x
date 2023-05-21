package header

import (
	"github.com/zalgonoise/x/audio/errs"
)

const (
	ErrDomain = errs.Domain("audio/wav/header")

	ErrShort   = errs.Kind("short")
	ErrEmpty   = errs.Kind("missing")
	ErrInvalid = errs.Kind("invalid")

	ErrNumChannels = errs.Entity("number of channels")
	ErrSampleRate  = errs.Entity("sample rate")
	ErrBitDepth    = errs.Entity("bit depth")
	ErrHeader      = errs.Entity("WAV header")
	ErrAudioFormat = errs.Entity("audio format")
	ErrDataBuffer  = errs.Entity("data buffer")
)

var (
	ErrEmptyHeader        = errs.New(ErrDomain, ErrEmpty, ErrHeader)
	ErrInvalidNumChannels = errs.New(ErrDomain, ErrInvalid, ErrNumChannels)
	ErrInvalidSampleRate  = errs.New(ErrDomain, ErrInvalid, ErrSampleRate)
	ErrInvalidBitDepth    = errs.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader      = errs.New(ErrDomain, ErrInvalid, ErrHeader)
	ErrInvalidAudioFormat = errs.New(ErrDomain, ErrInvalid, ErrAudioFormat)
	ErrShortDataBuffer    = errs.New(ErrDomain, ErrShort, ErrDataBuffer)
)
