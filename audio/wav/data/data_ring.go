package data

import (
	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data/conv"
	"github.com/zalgonoise/x/audio/wav/data/header"
	"io"
	"time"
)

const minRingSize = 64

// DataRing is a general-purpose chunk for audio data
type DataRing struct {
	ChunkHeader *header.Header
	Data        *gbuf.RingBuffer[float64]
	Depth       uint16
	Converter   Converter

	byteSize int
}

func (d *DataRing) Write(buf []byte) (n int, err error) {
	return d.Data.Write(d.Converter.Parse(buf))
}

func (d *DataRing) Read(buf []byte) (n int, err error) {
	return copy(buf, d.Converter.Bytes(d.Data.Value())), nil
}

func (d *DataRing) ReadFrom(b io.Reader) (n int64, err error) {
	buf := gbuf.NewRingFilter[byte](
		d.Data.Cap()*int(d.Depth/byteSize),
		func(data []byte) error {
			_, err = d.Data.Write(d.Converter.Parse(data))

			return err
		})

	return buf.ReadFrom(b)
}

func (d *DataRing) Parse(buf []byte) {
	_, _ = d.Data.Write(d.Converter.Parse(buf))
}

func (d *DataRing) ParseFloat(buf []float64) {
	_, _ = d.Data.Write(buf)
}

func (d *DataRing) Bytes() []byte {
	return d.Converter.Bytes(d.Data.Value())
}

func (d *DataRing) Header() *header.Header {
	return d.ChunkHeader
}

func (d *DataRing) BitDepth() uint16 {
	return d.Depth
}

func (d *DataRing) Reset() {
	d.Data.Reset()
}

func (d *DataRing) Value() []int {
	return d.Converter.Value(d.Data.Value())
}

func (d *DataRing) Float() []float64 {
	return d.Data.Value()
}

func (d *DataRing) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	max := d.Data.Cap()
	size := int(float64(sampleRate) * float64(dur) / float64(time.Second))

	if size > max {
		size = max
	}

	buf := make([]float64, size)

	oscillator := osc.NewOscillator(waveType)
	if oscillator == nil {
		return
	}

	oscillator(buf, freq, int(d.Depth), sampleRate)

	_, _ = d.Data.Write(buf)
}

func (d *DataRing) Apply(filters ...FilterFunc) {
	data := d.Data.Value()

	for i := range filters {
		filters[i](data)
	}

	_, _ = d.Data.Write(data)
}

func NewPCMDataRing(size int, bitDepth uint16, h *header.Header) *DataRing {
	if h == nil {
		h = header.NewData()
	}

	if size < minRingSize {
		size = minRingSize
	}

	switch bitDepth {
	case bitDepth8:
		return &DataRing{
			ChunkHeader: h,
			Data:        gbuf.NewRingBuffer[float64](size),
			Depth:       bitDepth,
			Converter:   conv.Conv8Bit{},
			byteSize:    size8,
		}
	case bitDepth16:
		return &DataRing{
			ChunkHeader: h,
			Data:        gbuf.NewRingBuffer[float64](size / (int(bitDepth) / byteSize)),
			Depth:       bitDepth,
			Converter:   conv.Conv16Bit{},
			byteSize:    size16,
		}
	case bitDepth24:
		return &DataRing{
			ChunkHeader: h,
			Data:        gbuf.NewRingBuffer[float64](size / (int(bitDepth) / byteSize)),
			Depth:       bitDepth,
			Converter:   conv.Conv24Bit{},
			byteSize:    size24,
		}
	case bitDepth32:
		return &DataRing{
			ChunkHeader: h,
			Data:        gbuf.NewRingBuffer[float64](size / (int(bitDepth) / byteSize)),
			Depth:       bitDepth,
			Converter:   conv.Conv32Bit{},
			byteSize:    size32,
		}
	default:
		return nil
	}
}
