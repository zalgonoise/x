package data

import (
	"bytes"
	"io"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

// Junk is a Chunk used for storing "junk"-ID subchunk data.
type Junk struct {
	ChunkHeader *Header
	Data        []byte
	Depth       uint16

	written int
}

// Write implements the io.Writer interface.
//
// It allows to grow the Junk's data with the input `buf` bytes, returning the number of
// bytes consumed and an error.
func (d *Junk) Write(buf []byte) (n int, err error) {
	d.Parse(buf)

	return len(buf), nil
}

// Read implements the io.Reader interface.
//
// It writes the data of the Junk into the input `buf`, returning the number of bytes read
// and an error.
func (d *Junk) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Data), nil
}

// ReadFrom implements the io.ReaderFrom interface.
//
// It reads exactly the chunk's declared size (reassembling short reads, and growing the buffer as data arrives instead
// of trusting the header for an allocation), then skips the RIFF pad byte of an odd-sized chunk. A chunk is read once:
// later calls are no-ops.
func (d *Junk) ReadFrom(r io.Reader) (n int64, err error) {
	switch {
	case d.ChunkHeader == nil:
		return 0, ErrMissingHeader
	case d.ChunkHeader.Subchunk2Size == 0, d.written > 0:
		return 0, nil
	}

	var buf bytes.Buffer

	n, err = readBounded(r, d.ChunkHeader, &buf)
	d.Data = buf.Bytes()
	d.written = len(d.Data)

	return n, err
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes.
func (d *Junk) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = buf

		if d.ChunkHeader.Subchunk2Size == 0 {
			d.ChunkHeader.Subchunk2Size = chunkSize(len(buf))
		}

		return
	}

	d.Data = append(d.Data, buf...)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data.
func (d *Junk) ParseFloat(_ []float64) {}

// Bytes will return a slice of bytes with the encoded PCM buffer.
func (d *Junk) Bytes() []byte {
	return d.Data
}

// Header returns the ChunkHeader of the Junk, with Subchunk2Size set to the size of the data it holds once any data
// was read or written (so a re-encode never declares more, or fewer, bytes than it writes).
func (d *Junk) Header() *Header {
	if d.Data != nil {
		d.ChunkHeader.Subchunk2Size = uint32(len(d.Data)) //nolint:gosec // bounded by the decoded input
	}

	return d.ChunkHeader
}

// BitDepth returns the bit depth of the Junk.
func (d *Junk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Junk.
func (d *Junk) Reset() {
	d.Data = make([]byte, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the Chunk, as a slice of int.
func (d *Junk) Value() []int { return to[byte, int](d.Data) }

// Float returns the PCM audio buffer from the Chunk, as a slice of float64.
func (d *Junk) Float() []float64 {
	return nil
}

// Generate creates a wave of the given form, frequency and duration within this Junk.
func (d *Junk) Generate(_ osc.Type, _, _ int, _ time.Duration) {}

// Apply transforms the floating-point audio data with each FilterFunc in `filters`.
func (d *Junk) Apply(_ ...FilterFunc) {}

// NewJunk creates a Junk with the input `subchunk` ChunkHeader, or with a default one if nil.
func NewJunk(h *Header) *Junk {
	if h == nil {
		h = NewJunkHeader()
	}

	return &Junk{
		ChunkHeader: h,
		Depth:       0,
	}
}
