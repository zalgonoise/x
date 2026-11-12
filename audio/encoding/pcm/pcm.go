// Package pcm converts between encoded audio samples and the module's normalised float64 samples.
//
// Normalisation follows the libsndfile / ffmpeg convention: an n-bit integer sample v decodes to v / 2^(n-1), so the
// range is [-1, 1 - 2^-(n-1)]; encoding scales by 2^(n-1), rounds to the nearest integer and clips to the integer
// range (NaN encodes as silence). Decoding then encoding any integer sample is exact.
//
// Codecs are safe for concurrent use and never retain the slices they are given.
package pcm

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/zalgonoise/x/audio"
)

// Codec converts between encoded samples and normalised float64 samples.
type Codec interface {
	// SampleSize returns the size of one encoded sample in bytes.
	SampleSize() int
	// Decode decodes min(len(dst), len(src)/SampleSize()) samples from src into dst and returns how many it decoded.
	// A trailing partial sample in src is ignored.
	Decode(dst []float64, src []byte) int
	// AppendEncode appends the encoding of every sample in src to dst and returns the extended slice.
	AppendEncode(dst []byte, src []float64) []byte
}

// WAV returns the codec RIFF/WAVE uses for a format: little-endian, with unsigned 8-bit PCM.
func WAV(f audio.Format) (Codec, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}

	switch f.Encoding {
	case audio.EncodingPCM:
		switch f.BitDepth {
		case 8:
			return Int{Bits: 8, Unsigned: true}, nil
		case 16, 24, 32:
			return Int{Bits: f.BitDepth, Order: binary.LittleEndian}, nil
		}
	case audio.EncodingFloat:
		switch f.BitDepth {
		case 32, 64:
			return Float{Bits: f.BitDepth, Order: binary.LittleEndian}, nil
		}
	case audio.EncodingALaw:
		if f.BitDepth == 8 {
			return ALaw{}, nil
		}
	case audio.EncodingMuLaw:
		if f.BitDepth == 8 {
			return MuLaw{}, nil
		}
	default: // EncodingUnknown is rejected by Validate
	}

	return nil, fmt.Errorf("%w: %s", audio.ErrUnsupportedFormat, f)
}

// Quantize converts a normalised sample to a bits-wide signed integer, for bits from 1 to 64: it scales by 2^(bits-1),
// rounds to the nearest integer and clips to [-2^(bits-1), 2^(bits-1)-1]. NaN becomes 0. Other widths panic.
func Quantize(f float64, bits int) int64 {
	if bits < 1 || bits > 64 {
		panic(fmt.Sprintf("pcm: Quantize to %d bits", bits))
	}

	if math.IsNaN(f) {
		return 0
	}

	scale := math.Ldexp(1, bits-1)
	v := math.Round(f * scale)
	maxV := int64(math.MaxInt64 >> (64 - bits))

	switch {
	case v >= scale: // 2^(bits-1) itself does not fit
		return maxV
	case v <= -scale:
		return -maxV - 1
	default: // |v| < 2^(bits-1) <= 2^63: the conversion is exact
		return int64(v)
	}
}

// Normalize converts a bits-wide signed integer sample to a normalised float64 (v / 2^(bits-1)).
func Normalize(v int64, bits int) float64 {
	return math.Ldexp(float64(v), -(bits - 1))
}
