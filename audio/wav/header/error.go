package header

import (
	"github.com/zalgonoise/x/audio/err"
)

const (
	ErrDomain = err.Domain("audio/wav/header")

	ErrEmpty   = err.Kind("missing")
	ErrInvalid = err.Kind("invalid")
	ErrShort   = err.Kind("short")

	ErrNumChannels = err.Entity("number of channels")
	ErrSampleRate  = err.Entity("sample rate")
	ErrBitDepth    = err.Entity("bit depth")
	ErrHeader      = err.Entity("WAV header")
	ErrDataBuffer  = err.Entity("data buffer")
	ErrAudioFormat = err.Entity("audio format")
)

var (
	ErrInvalidNumChannels = err.New(ErrDomain, ErrInvalid, ErrNumChannels)
	ErrInvalidSampleRate  = err.New(ErrDomain, ErrInvalid, ErrSampleRate)
	ErrInvalidBitDepth    = err.New(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader      = err.New(ErrDomain, ErrInvalid, ErrHeader)
	ErrShortDataBuffer    = err.New(ErrDomain, ErrShort, ErrDataBuffer)
	ErrMissingHeader      = err.New(ErrDomain, ErrEmpty, ErrHeader)
	ErrMissingDataBuffer  = err.New(ErrDomain, ErrInvalid, ErrDataBuffer)
	ErrInvalidAudioFormat = err.New(ErrDomain, ErrInvalid, ErrAudioFormat)
)
