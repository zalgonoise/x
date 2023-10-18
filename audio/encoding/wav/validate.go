package wav

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/x/audio/validation"
	"github.com/zalgonoise/x/errs"
)

const (
	SampleRate44100  uint32 = 44100
	SampleRate48000  uint32 = 48000
	SampleRate88200  uint32 = 88200
	SampleRate96000  uint32 = 96000
	SampleRate176400 uint32 = 176400
	SampleRate192000 uint32 = 192000
)

const (
	ChannelsMono uint16 = iota + 1
	ChannelsStereo
)

const (
	BitDepth8 uint16 = 8 * (iota + 1)
	BitDepth16
	BitDepth24
	BitDepth32
)

const (
	UnsetFormat uint16 = iota
	PCMFormat
	_
	FloatFormat
)

const (
	ErrDomain = errs.Domain("audio/wav")

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
	ErrEmptyHeader        = errs.WithDomain(ErrDomain, ErrEmpty, ErrHeader)
	ErrInvalidNumChannels = errs.WithDomain(ErrDomain, ErrInvalid, ErrNumChannels)
	ErrInvalidSampleRate  = errs.WithDomain(ErrDomain, ErrInvalid, ErrSampleRate)
	ErrInvalidBitDepth    = errs.WithDomain(ErrDomain, ErrInvalid, ErrBitDepth)
	ErrInvalidHeader      = errs.WithDomain(ErrDomain, ErrInvalid, ErrHeader)
	ErrInvalidAudioFormat = errs.WithDomain(ErrDomain, ErrInvalid, ErrAudioFormat)
	ErrShortDataBuffer    = errs.WithDomain(ErrDomain, ErrShort, ErrDataBuffer)
)

var headerValidator = validation.Register[*Header](
	validateChunkID,
	validateFormat,
	validateSampleRate,
	validateBitDepth,
	validateNumChannels,
	validateAudioFormat,
)

// ValidateHeader verifies that the input Header `h` is not nil and that it is valid
func ValidateHeader(h *Header) error {
	if h == nil {
		return ErrEmptyHeader
	}

	return headerValidator.Validate(h)
}

// Check confirms whether the input bytes are likely to be a header
// with the least operations possible
func Check(buf []byte) bool {
	if len(buf) < Size {
		return false
	}

	return bytes.Equal(buf[:chunkIDEnd], defaultChunkID[:]) &&
		bytes.Equal(buf[formatOffset:subChunkIDEnd], formatAndSubchunkID)
}

func validateChunkID(h *Header) error {
	if string(h.ChunkID[:]) != string(defaultChunkID[:]) {
		return fmt.Errorf("%w: ChunkID %s", ErrInvalidHeader, string(h.ChunkID[:]))
	}

	return nil
}

func validateFormat(h *Header) error {
	if string(h.Format[:]) != string(defaultFormat[:]) {
		return fmt.Errorf("%w: Format %s", ErrInvalidHeader, string(h.Format[:]))
	}

	return nil
}

func validateSampleRate(h *Header) error {
	switch h.SampleRate {
	case SampleRate44100, SampleRate48000, SampleRate88200, SampleRate96000, SampleRate176400, SampleRate192000:
		return nil
	default:
		return fmt.Errorf("%w: SampleRate %d", ErrInvalidSampleRate, h.SampleRate)
	}
}

func validateBitDepth(h *Header) error {
	switch h.BitsPerSample {
	case BitDepth8, BitDepth16, BitDepth24, BitDepth32:
		return nil
	default:
		return fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, h.BitsPerSample)
	}
}

func validateNumChannels(h *Header) error {
	switch h.NumChannels {
	case ChannelsMono, ChannelsStereo:
		return nil
	default:
		return fmt.Errorf("%w: NumChannels %d", ErrInvalidNumChannels, h.NumChannels)
	}
}

func validateAudioFormat(h *Header) error {
	switch h.AudioFormat {
	case PCMFormat, FloatFormat:
		return nil
	default:
		return fmt.Errorf("%w: AudioFormat %d", ErrInvalidAudioFormat, h.AudioFormat)
	}
}
