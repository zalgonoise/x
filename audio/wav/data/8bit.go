package data

import (
	"unsafe"

	"github.com/zalgonoise/x/audio/wav/forms"
)

const (
	maxInt8 float64 = 1<<7 - 1
	// minInt8 float64 = ^1<<6 + 1
)

// Chunk8bit is a Chunk used for 8 bit-depth PCM buffers
type Chunk8bit struct {
	*ChunkHeader
	Data  []int8
	Depth uint16
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *Chunk8bit) Parse(buf []byte) {
	if d.Data == nil {
		// fast cast to int8
		d.Data = *(*[]int8)(unsafe.Pointer(&buf))
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, *(*[]int8)(unsafe.Pointer(&buf))...)
}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *Chunk8bit) Bytes() []byte {
	return *(*[]byte)(unsafe.Pointer(&d.Data))
}

// Header returns the ChunkHeader of the Chunk
func (d *Chunk8bit) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the Chunk
func (d *Chunk8bit) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk
func (d *Chunk8bit) Reset() { d.Data = nil }

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *Chunk8bit) Value() []int { return to[int8, int](d.Data) }

func (d *Chunk8bit) Float() []float64 {
	return conv[int8, float64](
		d.Data, func(v int8) float64 {
			return float64(v) / maxInt8
		},
	)
}

func (d *Chunk8bit) Generate(formType forms.Type, freq, duration, sampleRate float64) {
	buffer := make([]int8, int(sampleRate*duration))
	fn := formFunc8bit(formType)
	if fn == nil {
		return
	}
	fn(buffer, freq, float64(d.Depth), sampleRate)

	if d.Data == nil {
		d.Data = buffer
		return
	}
	d.Data = append(d.Data, buffer...)
}

func formFunc8bit(typ forms.Type) forms.FormFunc[int8] {
	switch typ {
	case forms.SineWave:
		return forms.Sine[int8]
	default:
		return nil
	}
}
