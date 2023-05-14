package wav

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidNumChannels = NewError(ErrInvalid, ErrNumChannels)
	ErrInvalidSampleRate  = NewError(ErrInvalid, ErrSampleRate)
	ErrInvalidBitDepth    = NewError(ErrInvalid, ErrBitDepth)
	ErrInvalidHeader      = NewError(ErrInvalid, ErrHeader)
	ErrShortDataBuffer    = NewError(ErrShort, ErrDataBuffer)
	ErrMissingHeader      = NewError(ErrEmpty, ErrHeader)
	ErrMissingDataBuffer  = NewError(ErrInvalid, ErrDataBuffer)
	ErrInvalidAudioFormat = NewError(ErrInvalid, ErrAudioFormat)

	ErrEmpty   ErrorKind = EmptyError("missing")
	ErrInvalid ErrorKind = InvalidError("invalid")
	ErrShort   ErrorKind = ShortError("short")

	ErrNumChannels ErrorEntity = NumChannelsError("number of channels")
	ErrSampleRate  ErrorEntity = SampleRateError("sample rate")
	ErrBitDepth    ErrorEntity = BitDepthError("bit depth")
	ErrHeader      ErrorEntity = HeaderError("WAV header")
	ErrDataBuffer  ErrorEntity = DataBufferError("data buffer")
	ErrAudioFormat ErrorEntity = AudioFormatError("audio format")
)

type Error struct {
	ErrorKind
	ErrorEntity
	error string
}

func (e Error) Error() string {
	return e.error
}

func (e Error) Unwrap() error {
	return errors.Join(e.ErrorKind, e.ErrorEntity)
}

func NewError(kind ErrorKind, entity ErrorEntity) error {
	sb := new(strings.Builder)
	sb.WriteString("audio/wav: ")
	sb.WriteString(kind.Error())
	sb.WriteByte(' ')
	sb.WriteString(entity.Error())

	return Error{
		ErrorKind:   kind,
		ErrorEntity: entity,
		error:       sb.String(),
	}
}

func Errorf(kind ErrorKind, entity ErrorEntity, args ...any) error {
	var err string

	sb := new(strings.Builder)
	sb.WriteString("audio/wav: ")
	sb.WriteString(kind.Error())
	sb.WriteByte(' ')
	sb.WriteString(entity.Error())
	err = sb.String()

	if len(args) > 0 {
		err = fmt.Sprint(err, args)
	}

	return Error{
		ErrorKind:   kind,
		ErrorEntity: entity,
		error:       err,
	}
}

type ErrorKind error
type ErrorEntity error

type EmptyError string

func (e EmptyError) Error() string { return (string)(e) }

type InvalidError string

func (e InvalidError) Error() string { return (string)(e) }

type ShortError string

func (e ShortError) Error() string { return (string)(e) }

type NumChannelsError string

func (e NumChannelsError) Error() string { return (string)(e) }

type SampleRateError string

func (e SampleRateError) Error() string { return (string)(e) }

type BitDepthError string

func (e BitDepthError) Error() string { return (string)(e) }

type HeaderError string

func (e HeaderError) Error() string { return (string)(e) }

type DataBufferError string

func (e DataBufferError) Error() string { return (string)(e) }

type AudioFormatError string

func (e AudioFormatError) Error() string { return (string)(e) }
