package data

import "unsafe"

type Chunk16bit struct {
	*ChunkHeader
	Data  []int16
	Depth uint16
}

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

func (d *Chunk16bit) Generate() []byte {
	data := make([]byte, len(d.Data)*2)
	for i := range d.Data {
		append2Bytes(i, data, *(*[2]byte)(unsafe.Pointer(&d.Data[i])))
	}

	return data
}

func (d *Chunk16bit) Header() *ChunkHeader { return d.ChunkHeader }
func (d *Chunk16bit) BitDepth() uint16     { return d.Depth }
func (d *Chunk16bit) Reset()               { d.Data = nil }
func (d *Chunk16bit) Value() []int         { return to[int16, int](d.Data) }
