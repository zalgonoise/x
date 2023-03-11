package data

import (
	"unsafe"
)

// Chunk24bit is a Chunk used for 24 bit-depth PCM buffers
type Chunk24bit struct {
	*ChunkHeader
	Data  []int32
	Depth uint16
}

// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
// from raw bytes
func (d *Chunk24bit) Parse(buf []byte) {
	if d.Data == nil {
		buf32bit := copy24to32(buf)
		d.Data = *(*[]int32)(unsafe.Pointer(&buf32bit))
		d.Data = d.Data[:len(buf32bit)/4]
		for i := range d.Data {
			if d.Data[i]&0x00800000 != 0 {
				d.Data[i] |= ^0xffffff // handle signed integers
			}
		}

		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}

	buf32bit := copy24to32(buf)
	newData := *(*[]int32)(unsafe.Pointer(&buf32bit))
	newData = newData[:len(buf32bit)/4]
	for i := range newData {
		if newData[i]&0x00800000 != 0 {
			newData[i] |= ^0xffffff // handle signed integers
		}
	}

	d.Data = append(d.Data, newData...)
}

// Generate will return a slice of bytes with the encoded PCM buffer
func (d *Chunk24bit) Generate() []byte {
	data := make([]byte, len(d.Data)*3)
	for i := range d.Data {
		append3Bytes(i, data, *(*[3]byte)(unsafe.Pointer(&d.Data[i])))
	}
	return data
}

// Header returns the ChunkHeader of the Chunk
func (d *Chunk24bit) Header() *ChunkHeader { return d.ChunkHeader }

// BitDepth returns the bit depth of the Chunk
func (d *Chunk24bit) BitDepth() uint16 { return d.Depth }

// Reset clears the data stored in the Chunk
func (d *Chunk24bit) Reset() { d.Data = nil }

// Value returns the PCM audio buffer from the Chunk, as a slice of int
func (d *Chunk24bit) Value() []int { return to[int32, int](d.Data) }
