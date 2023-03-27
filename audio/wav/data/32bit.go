package data

import (
	"time"
	"unsafe"

	"github.com/zalgonoise/x/audio/wav/osc"
)

const (
	maxInt32 float64 = 1<<31 - 1
	// minInt32 float64 = ^1<<30 + 1
)

// Chunk32bit is a Chunk used for 32 bit-depth PCM buffers
type Chunk32bit struct {
	*ChunkHeader
	Data  []int32
	Depth uint16
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *Chunk32bit) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = *(*[]int32)(unsafe.Pointer(&buf))
		d.Data = d.Data[:len(buf)/4]
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	newData := *(*[]int32)(unsafe.Pointer(&buf))
	d.Data = append(d.Data, newData[:len(buf)/4]...)
}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *Chunk32bit) Bytes() []byte {
	data := make([]byte, len(d.Data)*4)
	for i := range d.Data {
		append4Bytes(i, data, *(*[4]byte)(unsafe.Pointer(&d.Data[i])))
	}
	return data
}

// Header returns the ChunkHeader of the Chunk
func (d *Chunk32bit) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the Chunk
func (d *Chunk32bit) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk
func (d *Chunk32bit) Reset() { d.Data = nil }

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *Chunk32bit) Value() []int { return to[int32, int](d.Data) }

// Float returns the PCM audio buffer from the Chunk, as a slice of float64
func (d *Chunk32bit) Float() []float64 {
	return conv[int32, float64](
		d.Data, func(v int32) float64 {
			return float64(v) / maxInt32
		},
	)
}

// Generate creates a wave of the given form, frequency and duration within this Chunk
func (d *Chunk32bit) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buffer := make([]int32, int(float64(sampleRate)*float64(dur)/float64(time.Second)))
	fn := formFunc24and32bit(waveType)
	if fn == nil {
		return
	}
	fn(buffer, float64(freq), float64(d.Depth), float64(sampleRate))

	if d.Data == nil {
		d.Data = buffer
		return
	}
	d.Data = append(d.Data, buffer...)
}
