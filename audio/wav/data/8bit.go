package data

import (
	"unsafe"
)

type Chunk8bit struct {
	*ChunkHeader
	Data  []int8
	Depth uint16
}

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

func (d *Chunk8bit) Generate() []byte {
	return *(*[]byte)(unsafe.Pointer(&d.Data))
}

func (d *Chunk8bit) Header() *ChunkHeader { return d.ChunkHeader }
func (d *Chunk8bit) BitDepth() uint16     { return d.Depth }
func (d *Chunk8bit) Reset()               { d.Data = nil }
func (d *Chunk8bit) Value() []int         { return to[int8, int](d.Data) }
