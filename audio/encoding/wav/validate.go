package wav

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/valigator"
	"github.com/zalgonoise/x/errs"
)

const (
	SampleRate44100  uint32 = 44100
	SampleRate48000  uint32 = 48000
	SampleRate88200  uint32 = 88200
	SampleRate96000  uint32 = 96000
	SampleRate176400 uint32 = 176400
	SampleRate192000 uint32 = 192000

	// MinSampleRate and MaxSampleRate bound the sample rates accepted from a header.
	MinSampleRate uint32 = 1
	MaxSampleRate uint32 = 1_000_000
)

const (
	ChannelsMono uint16 = iota + 1
	ChannelsStereo

	// MaxChannels bounds the channel count accepted from a header.
	MaxChannels uint16 = 256
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

//nolint:gochecknoglobals // is initialized in the background for performance during runtime, as it is never mutated.
var headerValidator = valigator.New(
	validateChunkID,
	validateFormat,
	validateSubchunk1ID,
	validateSampleRate,
	validateBitDepth,
	validateNumChannels,
	validateAudioFormat,
	validateBlockAlign,
)

// ValidateHeader verifies that the input Header `h` is not nil and that it is valid.
func ValidateHeader(h *Header) error {
	if h == nil {
		return ErrEmptyHeader
	}

	return headerValidator.Validate(h)
}

// Check confirms whether the input bytes are likely to be a header
// with the least operations possible.
func Check(buf []byte) bool {
	if len(buf) < Size {
		return false
	}

	return bytes.Equal(buf[:chunkIDEnd], defaultChunkID[:]) &&
		bytes.Equal(buf[formatOffset:subChunkIDEnd], formatAndSubchunkID)
}

func validateChunkID(h *Header) error {
	if !bytes.Equal(h.ChunkID[:], defaultChunkID[:]) {
		return fmt.Errorf("%w: ChunkID %s", ErrInvalidHeader, string(h.ChunkID[:]))
	}

	return nil
}

func validateFormat(h *Header) error {
	if !bytes.Equal(h.Format[:], defaultFormat[:]) {
		return fmt.Errorf("%w: Format %s", ErrInvalidHeader, string(h.Format[:]))
	}

	return nil
}

func validateSubchunk1ID(h *Header) error {
	if !bytes.Equal(h.Subchunk1ID[:], defaultSubchunk1ID[:]) {
		return fmt.Errorf("%w: Subchunk1ID %q, want \"fmt \"", ErrInvalidHeader, string(h.Subchunk1ID[:]))
	}

	return nil
}

// validateSampleRate accepts any structurally sane sample rate (it used to accept only six common rates, rejecting
// valid 8k, 22.05k or 32k files).
func validateSampleRate(h *Header) error {
	if h.SampleRate < MinSampleRate || h.SampleRate > MaxSampleRate {
		return fmt.Errorf("%w: SampleRate %d (want %d..%d)", ErrInvalidSampleRate, h.SampleRate, MinSampleRate, MaxSampleRate)
	}

	return nil
}

func validateBitDepth(h *Header) error {
	switch h.BitsPerSample {
	case BitDepth8, BitDepth16, BitDepth24, BitDepth32:
		return nil
	default:
		return fmt.Errorf("%w: BitsPerSample %d", ErrInvalidBitDepth, h.BitsPerSample)
	}
}

// validateNumChannels accepts any channel count from 1 to MaxChannels (it used to accept only mono and stereo).
func validateNumChannels(h *Header) error {
	if h.NumChannels < ChannelsMono || h.NumChannels > MaxChannels {
		return fmt.Errorf("%w: NumChannels %d (want 1..%d)", ErrInvalidNumChannels, h.NumChannels, MaxChannels)
	}

	return nil
}

func validateAudioFormat(h *Header) error {
	switch h.AudioFormat {
	case PCMFormat:
		return nil
	case FloatFormat:
		// IEEE float samples are 32-bit here (64-bit is not supported yet); a "16-bit float" header cannot be decoded.
		if h.BitsPerSample != BitDepth32 {
			return fmt.Errorf("%w: BitsPerSample %d for IEEE float", ErrInvalidBitDepth, h.BitsPerSample)
		}

		return nil
	default:
		return fmt.Errorf("%w: AudioFormat %d", ErrInvalidAudioFormat, h.AudioFormat)
	}
}

// validateBlockAlign rejects headers whose frame size disagrees with the channel count and bit depth: decoding relies
// on BlockAlign to split frames, and a zero value would divide by zero.
func validateBlockAlign(h *Header) error {
	if want := h.NumChannels * h.BitsPerSample / byteSize; h.BlockAlign != want {
		return fmt.Errorf("%w: BlockAlign %d, want %d (NumChannels * BitsPerSample / 8)", ErrInvalidHeader, h.BlockAlign, want)
	}

	return nil
}
