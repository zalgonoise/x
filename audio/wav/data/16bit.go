package data

import (
	"time"
	"unsafe"

	"github.com/zalgonoise/x/audio/osc"
)

const (
	maxInt16 float64 = 1<<15 - 1
	// minInt16 float64 = ^1<<14 + 1
)

// Chunk16bit is a Chunk used for 16 bit-depth PCM buffers
type Chunk16bit struct {
	*ChunkHeader
	Data  []int16
	Depth uint16
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *Chunk16bit) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = *(*[]int16)(unsafe.Pointer(&buf))
		d.Data = d.Data[:len(buf)/2]
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	newData := *(*[]int16)(unsafe.Pointer(&buf))
	d.Data = append(d.Data, newData[:len(buf)/2]...)
}

// ParseFloat will consume the input float64 slice `buf`, to extract the PCM audio buffer
// from floating-point audio data
func (d *Chunk16bit) ParseFloat(buf []float64) {
	d.Data = conv(
		buf, func(f float64) int16 {
			return int16(f * maxInt16)
		},
	)
}

// Bytes will return a slice of bytes with the encoded PCM buffer
func (d *Chunk16bit) Bytes() []byte {
	data := make([]byte, len(d.Data)*2)
	for i := range d.Data {
		append2Bytes(i, data, *(*[2]byte)(unsafe.Pointer(&d.Data[i])))
	}

	return data
}

// Header returns the ChunkHeader of the Chunk
func (d *Chunk16bit) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the Chunk
func (d *Chunk16bit) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk
func (d *Chunk16bit) Reset() { d.Data = nil }

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *Chunk16bit) Value() []int { return to[int16, int](d.Data) }

// Float returns the PCM audio buffer from the Chunk, as a slice of float64
func (d *Chunk16bit) Float() []float64 {
	return conv(
		d.Data, func(v int16) float64 {
			return float64(v) / maxInt16
		},
	)
}

// Generate creates a wave of the given form, frequency and duration within this Chunk
func (d *Chunk16bit) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buffer := make([]int16, int(float64(sampleRate)*float64(dur)/float64(time.Second)))
	fn := formFunc16bit(waveType)
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

func formFunc16bit(typ osc.Type) osc.Oscillator[int16] {
	switch typ {
	case osc.SineWave:
		return osc.Sine[int16]
	case osc.SquareWave:
		return osc.Square[int16]
	case osc.TriangleWave:
		return osc.Triangle[int16]
	case osc.SawtoothWave:
		return osc.Sawtooth[int16]
	default:
		return nil
	}
}
