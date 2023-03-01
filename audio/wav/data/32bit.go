package data

import "unsafe"

type Chunk32bit struct {
	*ChunkHeader
	Data  []int32
	Depth uint16
}

func (d *Chunk32bit) Parse(buf []byte) {
	if d.Data == nil {
		d.Data = *(*[]int32)(unsafe.Pointer(&buf))
		d.Data = d.Data[:len(buf)/4]
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	new := *(*[]int32)(unsafe.Pointer(&buf))
	d.Data = append(d.Data, new[:len(buf)/4]...)
}

func (d *Chunk32bit) Generate() []byte {
	data := make([]byte, len(d.Data)*4)
	for i := range d.Data {
		append4Bytes(i, data, *(*[4]byte)(unsafe.Pointer(&d.Data[i])))
	}
	return data
}

func (d *Chunk32bit) Header() *ChunkHeader { return d.ChunkHeader }
func (d *Chunk32bit) BitDepth() uint16     { return d.Depth }
func (d *Chunk32bit) Reset()               { d.Data = nil }
func (d *Chunk32bit) Value() []int         { return to[int32, int](d.Data) }
