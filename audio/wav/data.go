package wav

import (
	"io"
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data"
	datah "github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/header"
)

const (
	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32
)

// Chunk describes the behavior that a data chunk exposes, which involve
// reading and writing PCM audio buffers from / to bytes. Additionally, it
// provides helper methods to fetch the chunk header, the bit depth, to reset it
// and also to retrieve the PCM buffer as a slice of int
type Chunk interface {
	// ReadWriter joins both io.Read and io.Write interfaces, as a Chunk is able to both
	// be written to or read from.
	io.ReadWriter
	// ReaderFrom extends the io.Writer implementation to be able to consume the audio data from
	// an io.Reader, as well
	io.ReaderFrom
	// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
	// from raw bytes
	Parse(buf []byte)
	// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
	// from floating-point audio data
	ParseFloat(buf []float64)
	// Bytes will return a slice of bytes with the encoded PCM buffer
	Bytes() []byte
	// Header returns the ChunkHeader of the Chunk
	Header() *datah.Header
	// BitDepth returns the bit depth of the Chunk
	BitDepth() uint16
	// Reset clears the data stored in the Chunk
	Reset()
	// Value returns the PCM audio buffer from the Chunk, as a slice of int
	Value() []int
	// Float returns the PCM audio buffer from the Chunk, as a slice of float64
	Float() []float64
	// Generate creates a wave of the given form, frequency and duration within this Chunk
	Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration)
	// Apply transforms the floating-point audio data with each FilterFunc in `filters`
	Apply(filters ...data.FilterFunc)
}

// NewChunk is a factory for Chunk interfaces.
//
// Chunk are interfaces wrapping different types, based on the
// bit depth `bitDepth` value. These objects will store slices of integers of
// different sizes (int8, int16, int32, and bytes for "junk"), and the
// Chunk interface exposes the needed methods to work seamlessly with those
// different data types
//
// Note: I wanted a cleaner approach to this using generics and type constraints,
// but I was getting nowhere meaningful; and ended up breaking at a certain point
// due to the way that Go handles a slice of a type and its conversions to a different type
func NewChunk(h *datah.Header, bitDepth, format uint16) Chunk {
	if h != nil && string(h.Subchunk2ID[:]) == datah.JunkIDString {
		return data.NewJunkChunk(h)
	}

	switch format {
	case header.UnsetFormat:
		fallthrough
	case header.PCMFormat:
		switch bitDepth {
		case 0:
			return data.NewJunkChunk(h)
		case bitDepth8, bitDepth16, bitDepth24, bitDepth32:
			return data.NewPCMDataChunk(bitDepth, h)
		default:
			return nil
		}
	case header.FloatFormat:
		return data.NewFloatDataChunk(bitDepth, h)
	default:
		return nil
	}
}
