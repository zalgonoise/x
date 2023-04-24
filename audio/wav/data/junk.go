package data

import (
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

// ChunkJunk is a Chunk used for storing "junk"-ID subchunk data
type ChunkJunk struct {
	*ChunkHeader
	Data  []byte
	Depth uint16
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *ChunkJunk) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = buf
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, buf...)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data
func (d *ChunkJunk) ParseFloat(_ []float64) {}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *ChunkJunk) Bytes() []byte {
	return d.Data
}

// Header returns the ChunkHeader of the Chunk
func (d *ChunkJunk) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the Chunk
func (d *ChunkJunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk
func (d *ChunkJunk) Reset() {
	d.Data = make([]byte, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *ChunkJunk) Value() []int { return to[byte, int](d.Data) }

// Float returns the PCM audio buffer from the Chunk, as a slice of float64
func (d *ChunkJunk) Float() []float64 {
	return nil
}

// Generate creates a wave of the given form, frequency and duration within this Chunk
func (d *ChunkJunk) Generate(_ osc.Type, _, _ int, _ time.Duration) {}
