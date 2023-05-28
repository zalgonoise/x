package data

import (
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data/header"
)

const (
	dataChunkBaseLen = 512

	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32
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
	raw         []byte
}

// FilterFunc is a function that applies a transformation to a floating-point audio buffer
type FilterFunc func([]float64)

func (d *DataChunk) growChunkSize(v uint32) {
	d.ChunkHeader.Subchunk2Size += v
}

func (d *DataChunk) setChunkSize(v uint32) {
	d.ChunkHeader.Subchunk2Size = v
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *DataChunk) Parse(buf []byte) {
	ln := uint32(len(buf))
	d.Data = nil

	if d.raw == nil {
		d.raw = buf
		d.setChunkSize(ln)
		return
	}

	d.raw = append(d.raw, buf...)
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
	if len(d.raw) > 0 {
		return d.raw
	}

	d.raw = d.Converter.Bytes(d.Data)
	d.setChunkSize(uint32(len(d.raw)))
	return d.raw
}

// Header returns the ChunkHeader of the DataChunk
func (d *DataChunk) Header() *header.Header { return d.ChunkHeader }

// BitDepth returns the bit depth of the DataChunk
func (d *DataChunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the DataChunk
func (d *DataChunk) Reset() {
	d.Data = make([]float64, 0, dataChunkBaseLen)
	d.raw = make([]byte, 0, dataChunkBaseLen)
	d.setChunkSize(0)
}

// Value returns the PCM audio buffer from the DataChunk, as a slice of int
func (d *DataChunk) Value() []int {
	if len(d.Data) > 0 {
		return d.Converter.Value(d.Data)
	}

	d.Data = d.Converter.Parse(d.raw)
	return d.Converter.Value(d.Data)
}

// Float returns the PCM audio buffer from the DataChunk, as a slice of float64
func (d *DataChunk) Float() []float64 {
	if len(d.Data) > 0 {
		return d.Data
	}

	d.Data = d.Converter.Parse(d.raw)
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

	bufBytes := d.Converter.Bytes(buf)
	ln := uint32(len(bufBytes))

	if d.raw == nil {
		d.raw = bufBytes
		d.setChunkSize(ln)
	} else {
		d.raw = append(d.raw, bufBytes...)
		d.growChunkSize(ln)
	}

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

		return newChunk, nil
	}

	copy(newChunk.Data, newChunk.Converter.Parse(d.raw))

	newChunk.ChunkHeader.Subchunk2Size = uint32(len(newChunk.Converter.Bytes(d.Data)))

	return newChunk, nil
}

// Apply transforms the floating-point audio data with each FilterFunc in `filters`
func (d *DataChunk) Apply(filters ...FilterFunc) {
	if len(d.Data) == 0 && len(d.raw) > 0 {
		d.Data = d.Converter.Parse(d.raw)
		d.raw = make([]byte, 0, dataChunkBaseLen)
	}

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
		}
	case bitDepth16:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv16Bit{},
		}
	case bitDepth24:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv24Bit{},
		}
	case bitDepth32:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth,
			Converter:   Conv32Bit{},
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
		}
	default:
		return &DataChunk{
			ChunkHeader: h,
			Depth:       bitDepth32,
			Converter:   ConvFloat{},
		}
	}
}
