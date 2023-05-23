package header

import (
	"fmt"

	"github.com/zalgonoise/x/audio/validation"
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

var headerValidator = validation.New[*Header](
	validateChunkID,
	validateFormat,
	validateSampleRate,
	validateBitDepth,
	validateNumChannels,
	validateAudioFormat,
)

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

// Validate verifies that the input Header `h` is not nil and that it is valid
func Validate(h *Header) error {
	if h == nil {
		return ErrEmptyHeader
	}

	return headerValidator.Validate(h)
}
