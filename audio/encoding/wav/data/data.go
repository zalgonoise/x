package data

import (
	"io"
	"time"

	"github.com/zalgonoise/x/audio/encoding/wav/data/conv"
	"github.com/zalgonoise/x/audio/osc"
)

const (
	dataChunkBaseLen = 1024

	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32
	bitDepth64 uint16 = 64

	byteSize = 8

	// unitDepth is the bit depth for which the osc oscillators scale their output by 2^(depth-1)-1 = 1, producing the
	// normalized [-1, 1] samples the Converters expect.
	unitDepth = 2
)

// Converter describes the behavior that a bit-depth converter should expose -- that is to encode / decode a bytes
// buffer, as well as converting PCM audio data as int values.
type Converter interface {
	// Parse consumes the input audio buffer, returning its floating point audio representation.
	Parse(buf []byte) []float64
	// Bytes consumes the input floating point audio buffer, returning its byte representation.
	Bytes(buf []float64) []byte
	// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
	Value(buf []float64) []int
}

// Chunk is a general-purpose chunk for audio data.
type Chunk struct {
	ChunkHeader *Header
	Data        []float64
	Depth       uint16
	Converter   Converter

	byteSize   int
	blockAlign int
}

// FilterFunc is a function that applies a transformation to a floating-point audio buffer.
type FilterFunc func([]float64)

// Write implements the io.Writer interface.
//
// It allows to grow the Chunk's audio data with the input `buf` bytes, returning the number of
// bytes consumed and an error.
func (d *Chunk) Write(buf []byte) (n int, err error) {
	ln := len(buf)
	n = ln - ln%d.byteSize
	d.Parse(buf[:n])

	return n, nil
}

// Read implements the io.Reader interface.
//
// It writes the audio data of the Chunk into the input `buf`, returning the number of bytes read
// and an error.
func (d *Chunk) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Bytes()), nil
}

// ReadFrom implements the io.ReaderFrom interface.
//
// It consumes the audio data from the input io.Reader, until the chunk's declared size or, for a size of 0 or
// 0xFFFFFFFF (live streams), until io.EOF. Short reads are reassembled, so samples are always decoded from whole frames,
// and the RIFF pad byte after an odd-sized chunk is consumed rather than decoded as audio.
func (d *Chunk) ReadFrom(r io.Reader) (n int64, err error) {
	limit := declaredSize(d.ChunkHeader)

	n, err = readFrames(r, d.frameSize(), defaultReadSize, limit, func(b []byte) error {
		d.Data = append(d.Data, d.Converter.Parse(b)...)

		return nil
	})
	if err != nil || limit == 0 {
		return n, err
	}

	pad, err := skipPad(r, d.ChunkHeader)

	return n + pad, err
}

// SetBlockAlign sets the frame size in bytes (the WAV header's BlockAlign), so ReadFrom delivers whole frames.
// A zero value falls back to one sample.
func (d *Chunk) SetBlockAlign(blockAlign uint16) { d.blockAlign = int(blockAlign) }

func (d *Chunk) frameSize() int {
	if d.blockAlign > 0 {
		return d.blockAlign
	}

	return d.byteSize
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes.
func (d *Chunk) Parse(buf []byte) {
	if len(buf) == 0 {
		return
	}

	if d.Data == nil {
		d.Data = d.Converter.Parse(buf)

		return
	}

	d.Data = append(d.Data, d.Converter.Parse(buf)...)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data.
func (d *Chunk) ParseFloat(buf []float64) {
	if d.Data == nil {
		d.Data = buf
		return
	}

	d.Data = append(d.Data, buf...)
}

// Bytes will return a slice of bytes with the encoded PCM buffer.
func (d *Chunk) Bytes() []byte {
	if len(d.Data) == 0 {
		return nil
	}

	return d.Converter.Bytes(d.Data)
}

// Header returns the ChunkHeader of the Chunk, with Subchunk2Size set to the size of the samples it currently holds
// (so a re-encode never declares more, or fewer, bytes than it writes).
func (d *Chunk) Header() *Header {
	d.ChunkHeader.Subchunk2Size = uint32(len(d.Data) * d.byteSize) //nolint:gosec // bounded by the decoded input

	return d.ChunkHeader
}

// BitDepth returns the bit depth of the Chunk.
func (d *Chunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk.
func (d *Chunk) Reset() {
	d.Data = make([]float64, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the Chunk, as a slice of int.
func (d *Chunk) Value() []int {
	if len(d.Data) == 0 {
		return nil
	}

	return d.Converter.Value(d.Data)
}

// Float returns the PCM audio buffer from the Chunk, as a slice of float64.
func (d *Chunk) Float() []float64 {
	return d.Data
}

// Generate creates a wave of the given form, frequency and duration within this Chunk.
func (d *Chunk) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buf := make([]float64, int(float64(sampleRate)*float64(dur)/float64(time.Second)))

	oscillator := osc.NewOscillator(waveType)
	if oscillator == nil {
		return
	}

	// samples are normalized floats: request a unit-amplitude wave rather than one scaled to the bit depth
	oscillator(buf, freq, unitDepth, sampleRate)

	if d.Data == nil {
		d.Data = buf
		return
	}

	d.Data = append(d.Data, buf...)
}

// Apply transforms the floating-point audio data with each FilterFunc in `filters`.
func (d *Chunk) Apply(filters ...FilterFunc) {
	for i := range filters {
		filters[i](d.Data)
	}
}

// NewPCMChunk creates a PCM Chunk with the appropriate Converter, from the input
// `bitDepth` and `subchunk`.
func NewPCMChunk(bitDepth uint16, h *Header) *Chunk {
	if h == nil {
		h = NewDataHeader()
	}

	switch bitDepth {
	case bitDepth8:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   conv.PCM8Bit{},
			byteSize:    size8,
		}
	case bitDepth16:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   conv.PCM16Bit{},
			byteSize:    size16,
		}
	case bitDepth24:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   conv.PCM24Bit{},
			byteSize:    size24,
		}
	case bitDepth32:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   conv.PCM32Bit{},
			byteSize:    size32,
		}
	default:
		return nil
	}
}

// NewFloatChunk creates a 32-bit Float Chunk with the appropriate Converter, from the input
// `bitDepth` and `subchunk`.
func NewFloatChunk(bitDepth uint16, h *Header) *Chunk {
	if h == nil {
		h = NewDataHeader()
	}

	switch bitDepth {
	case bitDepth64:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth64,
			Converter:   conv.Float64{},
			byteSize:    int(bitDepth) / byteSize,
		}
	default:
		return &Chunk{
			ChunkHeader: h,
			Depth:       bitDepth32,
			Converter:   conv.Float32{},
			byteSize:    size32,
		}
	}
}
