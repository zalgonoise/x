// Package conv adapts the encoding/pcm codecs to the data.Converter interface used by encoding/wav/data.
//
// Samples are normalised by 2^(n-1) (encoding/pcm): an n-bit sample v decodes to v / 2^(n-1), and encoding rounds and
// clips. This package is kept for the existing encoding/wav types; new code should use encoding/pcm directly.
package conv

import (
	"github.com/zalgonoise/x/audio/encoding/pcm"
)

// BitDepthTypes is a type constraint joining all the different
// data types used by the supported bit depths.
type BitDepthTypes interface {
	int8 | int16 | int32 | uint32 | uint64 | byte | int | float64 | float32
}

func convert[From, To BitDepthTypes, FromSlice ~[]From](from FromSlice, fn func(From) To) []To {
	to := make([]To, len(from))

	for i := range from {
		to[i] = fn(from[i])
	}

	return to
}

func parse(c pcm.Codec, buf []byte) []float64 {
	out := make([]float64, len(buf)/c.SampleSize())
	c.Decode(out, buf)

	return out
}

func encode(c pcm.Codec, buf []float64) []byte {
	return c.AppendEncode(make([]byte, 0, len(buf)*c.SampleSize()), buf)
}

func quantize(buf []float64, bits int) []int {
	return convert(buf, func(f float64) int { return int(pcm.Quantize(f, bits)) })
}

// floatValueBits is the integer range Float32 / Float64 Value scale to.
const floatValueBits = 32

//nolint:gochecknoglobals // immutable codec values
var (
	codec8   = pcm.Int{Bits: 8, Unsigned: true}
	codec16  = pcm.Int{Bits: 16}
	codec24  = pcm.Int{Bits: 24}
	codec32  = pcm.Int{Bits: 32}
	codecF32 = pcm.Float{Bits: 32}
	codecF64 = pcm.Float{Bits: 64}
)

// PCM8Bit is a 8bit audio Converter. 8-bit WAV PCM is unsigned (0x80 is silence).
type PCM8Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM8Bit) Parse(buf []byte) []float64 { return parse(codec8, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM8Bit) Bytes(buf []float64) []byte { return encode(codec8, buf) }

// Value consumes the input floating point audio buffer, returning its PCM audio values (-128..127) as a slice of int.
func (PCM8Bit) Value(buf []float64) []int { return quantize(buf, 8) }

// PCM16Bit is a 16bit audio Converter.
type PCM16Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM16Bit) Parse(buf []byte) []float64 { return parse(codec16, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM16Bit) Bytes(buf []float64) []byte { return encode(codec16, buf) }

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (PCM16Bit) Value(buf []float64) []int { return quantize(buf, 16) }

// PCM24Bit is a 24bit audio Converter.
type PCM24Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM24Bit) Parse(buf []byte) []float64 { return parse(codec24, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM24Bit) Bytes(buf []float64) []byte { return encode(codec24, buf) }

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (PCM24Bit) Value(buf []float64) []int { return quantize(buf, 24) }

// PCM32Bit is a 32bit audio Converter.
type PCM32Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM32Bit) Parse(buf []byte) []float64 { return parse(codec32, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM32Bit) Bytes(buf []float64) []byte { return encode(codec32, buf) }

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (PCM32Bit) Value(buf []float64) []int { return quantize(buf, 32) }

// Float32 is a 32bit IEEE Floating Point audio Converter.
type Float32 struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (Float32) Parse(buf []byte) []float64 { return parse(codecF32, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (Float32) Bytes(buf []float64) []byte { return encode(codecF32, buf) }

// Value consumes the input floating point audio buffer, returning its values scaled to the 32-bit integer range,
// rounded and clipped.
func (Float32) Value(buf []float64) []int { return quantize(buf, floatValueBits) }

// Float64 is a 64bit IEEE Floating Point audio Converter.
type Float64 struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (Float64) Parse(buf []byte) []float64 { return parse(codecF64, buf) }

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (Float64) Bytes(buf []float64) []byte { return encode(codecF64, buf) }

// Value consumes the input floating point audio buffer, returning its values scaled to the 32-bit integer range,
// rounded and clipped.
func (Float64) Value(buf []float64) []int { return quantize(buf, floatValueBits) }
