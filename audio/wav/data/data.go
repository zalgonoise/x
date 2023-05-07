package data

import (
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

const chunkHeaderSize = 8

type Converter interface {
	Parse(buf []byte) []float64
	Bytes(buf []float64) []byte
	Value(buf []float64) []int
}

type DataChunk struct {
	ChunkHeader *ChunkHeader
	Data        []float64
	Depth       uint16
	Converter   Converter
}

func (d *DataChunk) growChunkSize(v uint32) {
	switch d.ChunkHeader.Subchunk2Size {
	case 0:
		d.ChunkHeader.Subchunk2Size = v + chunkHeaderSize
	default:
		d.ChunkHeader.Subchunk2Size += v
	}
}

func (d *DataChunk) setChunkSize(v uint32) {
	if d.ChunkHeader.Subchunk2Size == 0 {
		d.ChunkHeader.Subchunk2Size = v + chunkHeaderSize
	}
}

func (d *DataChunk) Parse(buf []byte) {
	ln := uint32(len(buf))

	if d.Data == nil {
		d.Data = d.Converter.Parse(buf)
		d.setChunkSize(ln)
		return
	}

	d.Data = append(d.Data, d.Converter.Parse(buf)...)
	d.growChunkSize(ln)
}

func (d *DataChunk) ParseFloat(buf []float64) {
	ln := uint32(len(d.Converter.Bytes(buf)))

	if d.Data == nil {
		d.Data = buf
		d.setChunkSize(ln)
		return
	}

	d.Data = append(d.Data, buf...)
	d.growChunkSize(ln)
}

func (d *DataChunk) Bytes() []byte {
	return d.Converter.Bytes(d.Data)
}

// Header returns the ChunkHeader of the DataChunk
func (d *DataChunk) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the DataChunk
func (d *DataChunk) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the DataChunk
func (d *DataChunk) Reset() {
	d.Data = make([]float64, 0, dataChunkBaseLen)
}

// Value returns the PCM audio buffer from the DataChunk, as a slice of int
func (d *DataChunk) Value() []int { return d.Converter.Value(d.Data) }

// Float returns the PCM audio buffer from the DataChunk, as a slice of float64
func (d *DataChunk) Float() []float64 { return d.Data }

// Generate creates a wave of the given form, frequency and duration within this DataChunk
func (d *DataChunk) Generate(waveType osc.Type, freq, sampleRate int, dur time.Duration) {
	buf := make([]float64, int(float64(sampleRate)*float64(dur)/float64(time.Second)))

	oscillator := osc.NewOscillator(waveType)
	if oscillator == nil {
		return
	}

	oscillator(buf, freq, int(d.Depth), sampleRate)

	ln := uint32(len(d.Converter.Bytes(buf)))

	if d.Data == nil {
		d.Data = buf
		d.setChunkSize(ln)
		return
	}

	d.Data = append(d.Data, buf...)
	d.growChunkSize(ln)
}

func (d *DataChunk) SetBitDepth(bitDepth uint16) (*DataChunk, error) {
	newChunk := NewDataChunk(bitDepth, d.ChunkHeader)
	if newChunk == nil {
		return nil, ErrInvalidBitDepth
	}

	copy(newChunk.Data, d.Data)

	newChunk.ChunkHeader.Subchunk2Size = uint32(len(newChunk.Converter.Bytes(d.Data)))

	return newChunk, nil
}

func NewDataChunk(bitDepth uint16, subchunk *ChunkHeader) *DataChunk {
	switch bitDepth {
	case bitDepth8:
		return &DataChunk{
			ChunkHeader: subchunk,
			Depth:       bitDepth,
			Converter:   Conv8Bit{},
		}
	case bitDepth16:
		return &DataChunk{
			ChunkHeader: subchunk,
			Depth:       bitDepth,
			Converter:   Conv16Bit{},
		}
	case bitDepth24:
		return &DataChunk{
			ChunkHeader: subchunk,
			Depth:       bitDepth,
			Converter:   Conv24Bit{},
		}
	case bitDepth32:
		return &DataChunk{
			ChunkHeader: subchunk,
			Depth:       bitDepth,
			Converter:   Conv32Bit{},
		}
	default:
		return nil
	}
}
