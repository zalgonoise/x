package data

import (
	"io"
	"time"

	"github.com/zalgonoise/gbuf"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data/header"
)

const (
	dataChunkBaseLen = 1024

	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32

	byteSize = 8
)

// Converter describes the behavior that a bit-depth converter should expose -- that is to encode / decode a bytes
// buffer, as well as converting PCM audio data as int values
type Converter interface {
	// Parse consumes the input audio buffer, returning its floating point audio representation
	Parse(buf []byte) []float64
	// Bytes consumes the input floating point audio buffer, returning its byte representation
	Bytes(buf []float64) []byte
	// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
	Value(buf []float64) []int
}

// DataChunk is a general-purpose chunk for audio data
type DataChunk struct {
	ChunkHeader *header.Header
	Data        []float64
	Depth       uint16
	Converter   Converter

	byteSize  int
	blockSize int
}

// FilterFunc is a function that applies a transformation to a floating-point audio buffer
type FilterFunc func([]float64)

func (d *DataChunk) growChunkSize(v uint32) {
	d.ChunkHeader.Subchunk2Size += v
}

func (d *DataChunk) setChunkSize(v uint32) {
	d.ChunkHeader.Subchunk2Size = v
}

// Write implements the io.Writer interface
//
// It allows to grow the DataChunk's audio data with the input `buf` bytes, returning the number of
// bytes consumed and an error
func (d *DataChunk) Write(buf []byte) (n int, err error) {
	ln := len(buf)
	n = ln - ln%d.byteSize
	d.Parse(buf[:n])

	return n, nil
}

// Read implements the io.Reader interface
//
// It writes the audio data of the DataChunk into the input `buf`, returning the number of bytes read
// and an error
func (d *DataChunk) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Bytes()), nil
}

// ReadFrom implements the io.ReaderFrom interface
//
// It consumes the audio data from the input io.Reader
func (d *DataChunk) ReadFrom(b io.Reader) (n int64, err error) {
	if d.blockSize == 0 {
		d.blockSize = dataChunkBaseLen / d.byteSize
	}

	size := d.blockSize * d.byteSize

	dataBuf := gbuf.NewRingFilter[float64](d.blockSize,
		func(data []float64) error {
			if d.Data == nil {
				d.Data = data

				return nil
			}

			d.Data = append(d.Data, data...)

			return nil
		},
	)

	buf := gbuf.NewRingFilter[byte](size, func(data []byte) error {
		_, err = dataBuf.Write(d.Converter.Parse(data))

		return err
	})

	return buf.ReadFrom(b)
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *DataChunk) Parse(buf []byte) {
	ln := uint32(len(buf))

	if ln == 0 {
		return
	}

	if d.Data == nil {
		d.Data = d.Converter.Parse(buf)
		d.setChunkSize(ln)
		return
	}

	d.Data = append(d.Data, d.Converter.Parse(buf)...)
	d.growChunkSize(ln)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data
func (d *DataChunk) ParseFloat(buf []float64) {
	if d.Data == nil {
		d.Data = buf
		return
	}

	d.Data = append(d.Data, buf...)
}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *DataChunk) Bytes() []byte {
	if len(d.Data) == 0 {
		return nil
	}

	return d.Converter.Bytes(d.Data)
}

// Header returns the ChunkHeader of the DataChunk
func (d *DataChunk) Header() *header.Header {
	if d.ChunkHeader.Subchunk2Size == 0 {
		d.ChunkHeader.Subchunk2Size = uint32(len(d.Data) * (int(d.Depth) / 8))
	}
	return d.ChunkHeader
}

// BitDepth returns the bit depth of the DataChunk
func (d *DataChunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the DataChunk
func (d *DataChunk) Reset() {
	d.Data = make([]float64, 0, dataChunkBaseLen)
	d.setChunkSize(0)
}

// Value returns the PCM audio buffer from the DataChunk, as a slice of int
func (d *DataChunk) Value() []int {
	if len(d.Data) == 0 {
		return nil
	}

	return d.Converter.Value(d.Data)
}

// Float returns the PCM audio buffer from the DataChunk, as a slice of float64
func (d *DataChunk) Float() []float64 {
	return d.Data
}

// Generate creates a wave of the given form, frequency and duration within this DataChunk
func (d *DataChunk) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buf := make([]float64, int(float64(sampleRate)*float64(dur)/float64(time.Second)))

	oscillator := osc.NewOscillator(waveType)
	if oscillator == nil {
		return
	}

	oscillator(buf, freq, int(d.Depth), sampleRate)

	if d.Data == nil {
		d.Data = buf
		return
	}

	d.Data = append(d.Data, buf...)
}

// SetBitDepth returns a new DataChunk with the input `bitDepth`'s converter, or
// an error if invalid. The new DataChunk retains any PCM data it contains, as a copy.
func (d *DataChunk) SetBitDepth(bitDepth uint16) (*DataChunk, error) {
	newChunk := NewPCMDataChunk(bitDepth, d.ChunkHeader)
	if newChunk == nil {
		return nil, ErrInvalidBitDepth
	}

	if len(d.Data) > 0 {
		copy(newChunk.Data, d.Data)

		newChunk.ChunkHeader.Subchunk2Size = uint32(len(newChunk.Converter.Bytes(d.Data)))
	}

	return newChunk, nil
}

// SetBufferSize delimits the size of the buffer, so that an audio stream keeps reusing the
// same pre-allocated buffer
func (d *DataChunk) SetBufferSize(size int) {
	if size < dataChunkBaseLen {
		return
	}

	d.blockSize = size
}

// Apply transforms the floating-point audio data with each FilterFunc in `filters`
func (d *DataChunk) Apply(filters ...FilterFunc) {
	for i := range filters {
		filters[i](d.Data)
	}
}

// NewPCMDataChunk creates a PCM DataChunk with the appropriate Converter, from the input
// `bitDepth` and `subchunk`
func NewPCMDataChunk(bitDepth uint16, h *header.Header) *DataChunk {
	if h == nil {
		h = header.NewData()
	}

	switch bitDepth {
	case bitDepth8:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv8Bit{},
			byteSize:    size8,
		}
	case bitDepth16:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv16Bit{},
			byteSize:    size16,
		}
	case bitDepth24:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv24Bit{},
			byteSize:    size24,
		}
	case bitDepth32:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv32Bit{},
			byteSize:    size32,
		}
	default:
		return nil
	}
}

// NewFloatDataChunk creates a 32-bit Float DataChunk with the appropriate Converter, from the input
// `bitDepth` and `subchunk`
func NewFloatDataChunk(bitDepth uint16, h *header.Header) *DataChunk {
	if h == nil {
		h = header.NewData()
	}

	switch bitDepth {
	case bitDepth8, bitDepth16, bitDepth24, bitDepth32:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   ConvFloat{},
			byteSize:    int(bitDepth) / byteSize,
		}
	default:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth32,
			Converter:   ConvFloat{},
			byteSize:    size32,
		}
	}
}
