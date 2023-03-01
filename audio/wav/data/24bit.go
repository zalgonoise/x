package data

import "unsafe"

type Chunk24bit struct {
	*ChunkHeader
	Data  []int32
	Depth uint16
}

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

func (d *Chunk24bit) Generate() []byte {
	data := make([]byte, len(d.Data)*3)
	for i := range d.Data {
		append3Bytes(i, data, *(*[3]byte)(unsafe.Pointer(&d.Data[i])))
	}
	return data
}

func (d *Chunk24bit) Header() *ChunkHeader { return d.ChunkHeader }
func (d *Chunk24bit) BitDepth() uint16     { return d.Depth }
func (d *Chunk24bit) Reset()               { d.Data = nil }
func (d *Chunk24bit) Value() []int         { return to[int32, int](d.Data) }
