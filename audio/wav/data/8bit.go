package data

import (
	"time"
	"unsafe"

	"github.com/zalgonoise/x/audio/osc"
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

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data
func (d *Chunk8bit) ParseFloat(buf []float64) {
	d.Data = conv(
		buf, func(f float64) int8 {
			return int8(f * maxInt8)
		},
	)
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
func (d *Chunk8bit) Reset() {
	d.Data = make([]int8, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *Chunk8bit) Value() []int { return to[int8, int](d.Data) }

// Float returns the PCM audio buffer from the Chunk, as a slice of float64
func (d *Chunk8bit) Float() []float64 {
	return conv(
		d.Data, func(v int8) float64 {
			return float64(v) / maxInt8
		},
	)
}

// Generate creates a wave of the given form, frequency and duration within this Chunk
func (d *Chunk8bit) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buffer := make([]int8, int(float64(sampleRate)*float64(dur)/float64(time.Second)))
	fn := formFunc8bit(waveType)
	if fn == nil {
		return
	}
	fn(buffer, freq, int(d.Depth), sampleRate)

	if d.Data == nil {
		d.Data = buffer
		return
	}
	d.Data = append(d.Data, buffer...)
}

func formFunc8bit(typ osc.Type) osc.Oscillator[int8] {
	switch typ {
	case osc.SineWave:
		return osc.Sine[int8]
	case osc.SquareWave:
		return osc.Square[int8]
	case osc.TriangleWave:
		return osc.Triangle[int8]
	case osc.SawtoothWave:
		return osc.Sawtooth[int8]
	default:
		return nil
	}
}
