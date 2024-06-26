package data

import (
	"io"
	"time"

	"github.com/zalgonoise/gbuf"

	"github.com/zalgonoise/x/audio/encoding/wav/data/conv"
	"github.com/zalgonoise/x/audio/osc"
)

const minRingSize = 64

// Ring is a general-purpose chunk for audio data.
type Ring struct {
	ChunkHeader *Header
	Data        *gbuf.RingFilter[float64]
	Depth       uint16
	Converter   Converter

	byteSize int
}

func (d *Ring) Write(buf []byte) (n int, err error) {
	return d.Data.Write(d.Converter.Parse(buf))
}

func (d *Ring) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Converter.Bytes(d.Data.Value())), nil
}

func (d *Ring) ReadFrom(b io.Reader) (n int64, err error) {
	buf := gbuf.NewRingFilter[byte](
		d.Data.Cap()*int(d.Depth/byteSize),
		func(data []byte) error {
			_, err = d.Data.Write(d.Converter.Parse(data))

			return err
		})

	return buf.ReadFrom(b)
}

func (d *Ring) Parse(buf []byte) {
	//nolint:errcheck // writing to the in-memory RingFilter should not raise any errors, and can be safely ignored.
	_, _ = d.Data.Write(d.Converter.Parse(buf))
}

func (d *Ring) ParseFloat(buf []float64) {
	//nolint:errcheck // writing to the in-memory RingFilter should not raise any errors, and can be safely ignored.
	_, _ = d.Data.Write(buf)
}

func (d *Ring) Bytes() []byte {
	return d.Converter.Bytes(d.Data.Value())
}

func (d *Ring) Header() *Header {
	d.ChunkHeader.Subchunk2Size = uint32(d.Data.Cap() * (int(d.Depth) / byteSize))

	return d.ChunkHeader
}

func (d *Ring) BitDepth() uint16 {
	return d.Depth
}

func (d *Ring) Reset() {
	d.Data.Reset()
}

func (d *Ring) Value() []int {
	return d.Converter.Value(d.Data.Value())
}

func (d *Ring) Float() []float64 {
	return d.Data.Value()
}

func (d *Ring) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	maximum := d.Data.Cap()
	size := int(float64(sampleRate) * float64(dur) / float64(time.Second))

	if size > maximum {
		size = maximum
	}

	buf := make([]float64, size)

	oscillator := osc.NewOscillator(waveType)
	if oscillator == nil {
		return
	}

	oscillator(buf, freq, int(d.Depth), sampleRate)

	//nolint:errcheck // writing to the in-memory RingFilter should not raise any errors, and can be safely ignored.
	_, _ = d.Data.Write(buf)
}

func (d *Ring) Apply(filters ...FilterFunc) {
	data := d.Data.Value()

	for i := range filters {
		filters[i](data)
	}

	//nolint:errcheck // writing to the in-memory RingFilter should not raise any errors, and can be safely ignored.
	_, _ = d.Data.Write(data)
}

// NewPCMRing creates a PCM Ring with the appropriate Converter, from the input
// `bitDepth` and `subchunk`, with the fixed buffer-size `size`.
func NewPCMRing(bitDepth uint16, h *Header, size int, proc func([]float64) error) *Ring {
	if h == nil {
		h = NewDataHeader()
	}

	if size < minRingSize {
		size = minRingSize
	}

	switch bitDepth {
	case bitDepth8:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size, proc),
			Depth:       bitDepth,
			Converter:   conv.PCM8Bit{},
			byteSize:    size8,
		}
	case bitDepth16:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size/(int(bitDepth)/byteSize), proc),
			Depth:       bitDepth,
			Converter:   conv.PCM16Bit{},
			byteSize:    size16,
		}
	case bitDepth24:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size/(int(bitDepth)/byteSize), proc),
			Depth:       bitDepth,
			Converter:   conv.PCM24Bit{},
			byteSize:    size24,
		}
	case bitDepth32:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size/(int(bitDepth)/byteSize), proc),
			Depth:       bitDepth,
			Converter:   conv.PCM32Bit{},
			byteSize:    size32,
		}
	default:
		return nil
	}
}

// NewFloatRing creates a 32-bit Float Ring with the appropriate Converter, from the input
// `bitDepth` and `subchunk`, with the fixed buffer-size `size`.
func NewFloatRing(bitDepth uint16, h *Header, size int, proc func([]float64) error) *Ring {
	if h == nil {
		h = NewDataHeader()
	}

	if size < minRingSize {
		size = minRingSize
	}

	switch bitDepth {
	case bitDepth64:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size/(int(bitDepth)/byteSize), proc),
			Depth:       bitDepth64,
			Converter:   conv.Float64{},
			byteSize:    int(bitDepth) / byteSize,
		}
	default:
		return &Ring{
			ChunkHeader: h,
			Data:        gbuf.NewRingFilter[float64](size/(int(bitDepth)/byteSize), proc),
			Depth:       bitDepth32,
			Converter:   conv.Float32{},
			byteSize:    size32,
		}
	}
}
