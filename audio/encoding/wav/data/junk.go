package data

import (
	"io"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

// JunkChunk is a DataChunk used for storing "junk"-ID subchunk data
type JunkChunk struct {
	ChunkHeader *Header
	Data        []byte
	Depth       uint16

	written int
}

// Write implements the io.Writer interface
//
// It allows to grow the JunkChunk's data with the input `buf` bytes, returning the number of
// bytes consumed and an error
func (d *JunkChunk) Write(buf []byte) (n int, err error) {
	d.Parse(buf)

	return len(buf), nil
}

// Read implements the io.Reader interface
//
// It writes the data of the JunkChunk into the input `buf`, returning the number of bytes read
// and an error
func (d *JunkChunk) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Data), nil
}

func (d *JunkChunk) ReadFrom(p io.Reader) (n int64, err error) {
	var size int

	switch {
	case d.ChunkHeader == nil:
		return 0, ErrMissingHeader
	case d.ChunkHeader.Subchunk2Size == 0:
		return n, nil
	default:
		size = int(d.ChunkHeader.Subchunk2Size)
	}

	if d.written == size {
		return 0, nil
	}

	if d.Data == nil {
		d.Data = make([]byte, size)
	}

	num, err := p.Read(d.Data[d.written:size])
	d.written = num

	return int64(num), err
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *JunkChunk) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = buf
		if d.ChunkHeader.Subchunk2Size == 0 {
			d.ChunkHeader.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, buf...)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data
func (d *JunkChunk) ParseFloat(_ []float64) {}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *JunkChunk) Bytes() []byte {
	return d.Data
}

// Header returns the ChunkHeader of the DataChunk
func (d *JunkChunk) Header() *Header { return d.ChunkHeader }

// BitDepth returns the bit depth of the DataChunk
func (d *JunkChunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the DataChunk
func (d *JunkChunk) Reset() {
	d.Data = make([]byte, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the DataChunk, as a slice of int
func (d *JunkChunk) Value() []int { return to[byte, int](d.Data) }

// Float returns the PCM audio buffer from the DataChunk, as a slice of float64
func (d *JunkChunk) Float() []float64 {
	return nil
}

// Generate creates a wave of the given form, frequency and duration within this DataChunk
func (d *JunkChunk) Generate(_ osc.Type, _, _ int, _ time.Duration) {}

// Apply transforms the floating-point audio data with each FilterFunc in `filters`
func (d *JunkChunk) Apply(_ ...FilterFunc) {}

// NewJunkChunk creates a JunkChunk with the input `subchunk` ChunkHeader, or with a default one if nil
func NewJunkChunk(h *Header) *JunkChunk {
	if h == nil {
		h = NewJunk()
	}

	return &JunkChunk{
		ChunkHeader: h,
		Depth:       0,
	}
}