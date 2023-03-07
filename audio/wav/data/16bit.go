package data

import "unsafe"

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
	new := *(*[]int16)(unsafe.Pointer(&buf))
	d.Data = append(d.Data, new[:len(buf)/2]...)
}

// Generate will return a slice of bytes with the encoded PCM buffer
func (d *Chunk16bit) Generate() []byte {
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
