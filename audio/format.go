package audio

import (
	"errors"
	"fmt"
	"time"
)

// Encoding is the way samples are stored.
type Encoding uint8

const (
	// EncodingUnknown is the zero value: not a usable encoding.
	EncodingUnknown Encoding = iota
	// EncodingPCM is linear PCM: signed integers, except 8-bit WAV data, which is unsigned with a 128 offset.
	EncodingPCM
	// EncodingFloat is IEEE-754 floating point (32 or 64 bits).
	EncodingFloat
	// EncodingALaw is ITU-T G.711 A-law (8 bits).
	EncodingALaw
	// EncodingMuLaw is ITU-T G.711 µ-law (8 bits).
	EncodingMuLaw
)

// Valid reports whether e is one of the defined encodings other than EncodingUnknown.
func (e Encoding) Valid() bool { return e > EncodingUnknown && e <= EncodingMuLaw }

// String returns the encoding's name.
func (e Encoding) String() string {
	switch e {
	case EncodingPCM:
		return "pcm"
	case EncodingFloat:
		return "float"
	case EncodingALaw:
		return "a-law"
	case EncodingMuLaw:
		return "mu-law"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(e))
	}
}

// Errors describing a Format.
var (
	// ErrInvalidFormat is returned for a Format that cannot describe any stream (zero channels, a bit depth that is not
	// a whole number of bytes, ...): the input is malformed.
	ErrInvalidFormat = errors.New("audio: invalid format")
	// ErrUnsupportedFormat is returned for a well-formed Format that no available codec handles.
	ErrUnsupportedFormat = errors.New("audio: unsupported format")
)

const bitsPerByte = 8

// Format describes a stream of audio samples, independently of its container.
type Format struct {
	SampleRate int      // frames per second
	Channels   int      // samples per frame
	BitDepth   int      // bits per sample in the container (a whole number of bytes)
	ValidBits  int      // significant bits within the container (WAVE_FORMAT_EXTENSIBLE); 0 means BitDepth
	Encoding   Encoding // how each sample is stored
	// ChannelMask maps channels to speaker positions (WAVE_FORMAT_EXTENSIBLE dwChannelMask); 0 means unspecified.
	ChannelMask uint32
}

// Validate reports whether the Format is structurally valid (ErrInvalidFormat otherwise). It does not check whether a
// codec exists for it.
func (f Format) Validate() error {
	switch {
	case f.SampleRate <= 0:
		return fmt.Errorf("%w: sample rate %d", ErrInvalidFormat, f.SampleRate)
	case f.Channels <= 0:
		return fmt.Errorf("%w: %d channels", ErrInvalidFormat, f.Channels)
	case f.BitDepth <= 0 || f.BitDepth%bitsPerByte != 0:
		return fmt.Errorf("%w: bit depth %d", ErrInvalidFormat, f.BitDepth)
	case f.ValidBits < 0 || f.ValidBits > f.BitDepth:
		return fmt.Errorf("%w: %d valid bits in a %d-bit container", ErrInvalidFormat, f.ValidBits, f.BitDepth)
	case !f.Encoding.Valid():
		return fmt.Errorf("%w: encoding %s", ErrInvalidFormat, f.Encoding)
	}

	return nil
}

// SampleSize is the size of one sample in bytes.
func (f Format) SampleSize() int { return f.BitDepth / bitsPerByte }

// BlockAlign is the size of one frame (one sample per channel) in bytes.
func (f Format) BlockAlign() int { return f.Channels * f.SampleSize() }

// ByteRate is the number of bytes per second of audio.
func (f Format) ByteRate() int { return f.SampleRate * f.BlockAlign() }

// Significant returns the number of significant bits per sample (ValidBits, or BitDepth when unset).
func (f Format) Significant() int {
	if f.ValidBits > 0 {
		return f.ValidBits
	}

	return f.BitDepth
}

// Duration returns how long the given number of frames lasts, rounded up to the nanosecond, so that
// Frames(Duration(n)) == n for any sample rate up to 1 GHz.
func (f Format) Duration(frames int64) time.Duration {
	if f.SampleRate <= 0 || frames <= 0 {
		return 0
	}

	rate := int64(f.SampleRate)
	secs, rest := frames/rate, frames%rate

	return time.Duration(secs)*time.Second + time.Duration((rest*int64(time.Second)+rate-1)/rate)
}

// Frames returns how many whole frames fit in the given duration.
func (f Format) Frames(d time.Duration) int64 {
	if d <= 0 || f.SampleRate <= 0 {
		return 0
	}

	secs, rest := int64(d/time.Second), int64(d%time.Second)

	return secs*int64(f.SampleRate) + rest*int64(f.SampleRate)/int64(time.Second)
}

// String describes the format, e.g. "pcm 16-bit 2ch 44100Hz".
func (f Format) String() string {
	bits := fmt.Sprintf("%d-bit", f.BitDepth)
	if f.ValidBits > 0 && f.ValidBits != f.BitDepth {
		bits = fmt.Sprintf("%d-in-%d-bit", f.ValidBits, f.BitDepth)
	}

	return fmt.Sprintf("%s %s %dch %dHz", f.Encoding, bits, f.Channels, f.SampleRate)
}
