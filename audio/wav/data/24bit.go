package data

import "unsafe"

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
		d.Data = conv(buf, 3, func(buf []byte) int32 {
			return decode24BitLE(buf)
		})
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}

	d.Data = append(d.Data, conv(buf, 3, func(buf []byte) int32 {
		return decode24BitLE(buf)
	})...)
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
