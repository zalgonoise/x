package wav

import (
	"encoding/binary"
	"unsafe"
)

type bitDepthTypes interface {
	int8 | int16 | int32 | byte
}

func conv[F, T bitDepthTypes](a []F, steps int, fn func([]F) T) []T {
	out := make([]T, len(a)/steps)
	for i, j := 0, 0; i+steps-1 < len(a); i, j = i+steps, j+1 {
		out[j] = fn(a[i : i+steps])
	}
	return out
}

type DataChunk interface {
	Parse(buf []byte, offset int)
	Generate() []byte
	Header() *SubChunk
	BitDepth() uint16
}

type DataChunkJunk struct {
	*SubChunk
	Data  []byte
	Depth uint16
}

func (d *DataChunkJunk) Parse(buf []byte, offset int) {
	if d.Data == nil {
		d.Data = buf
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, buf...)
}

func (d *DataChunkJunk) Generate() []byte {
	return d.Data
}

func (d *DataChunkJunk) Header() *SubChunk { return d.SubChunk }
func (d *DataChunkJunk) BitDepth() uint16  { return d.Depth }

type DataChunk8bit struct {
	*SubChunk
	Data  []int8
	Depth uint16
}

func (d *DataChunk8bit) Parse(buf []byte, offset int) {
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

func (d *DataChunk8bit) Generate() []byte {
	return *(*[]byte)(unsafe.Pointer(&d.Data))
}

func (d *DataChunk8bit) Header() *SubChunk { return d.SubChunk }
func (d *DataChunk8bit) BitDepth() uint16  { return d.Depth }

type DataChunk16bit struct {
	*SubChunk
	Data  []int16
	Depth uint16
}

func (d *DataChunk16bit) Parse(buf []byte, offset int) {
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

func (d *DataChunk16bit) Generate() []byte {
	n := len(d.Data)
	data := make([]byte, n*2)
	for i, j := 0, 0; i < n; i, j = i+1, j+2 {
		bin := *(*[2]byte)(unsafe.Pointer(&d.Data[i]))
		copy(data[j:j+2], bin[:])
	}

	return data
}

func (d *DataChunk16bit) Header() *SubChunk { return d.SubChunk }
func (d *DataChunk16bit) BitDepth() uint16  { return d.Depth }

type DataChunk24bit struct {
	*SubChunk
	Data  []int32
	Depth uint16
}

func (d *DataChunk24bit) Parse(buf []byte, offset int) {
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

func (d *DataChunk24bit) Generate() []byte {
	n := len(d.Data)
	data := make([]byte, n*3)
	for i, j := 0, 0; i < n; i, j = i+1, j+3 {
		bin := *(*[3]byte)(unsafe.Pointer(&d.Data[i]))
		copy(data[j:j+3], bin[:])
	}
	return data
}

func (d *DataChunk24bit) Header() *SubChunk { return d.SubChunk }
func (d *DataChunk24bit) BitDepth() uint16  { return d.Depth }

type DataChunk32bit struct {
	*SubChunk
	Data  []int32
	Depth uint16
}

func (d *DataChunk32bit) Parse(buf []byte, offset int) {
	if d.Data == nil {
		d.Data = conv(buf, 4, func(buf []byte) int32 {
			return int32(binary.LittleEndian.Uint32(buf))
		})
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, conv(buf, 4, func(buf []byte) int32 {
		return int32(binary.LittleEndian.Uint32(buf))
	})...)
}

func (d *DataChunk32bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*4)
	for i := 0; i < len(d.Data); i++ {
		data = binary.LittleEndian.AppendUint32(data, uint32(d.Data[i]))
	}
	return data
}

func (d *DataChunk32bit) Header() *SubChunk { return d.SubChunk }
func (d *DataChunk32bit) BitDepth() uint16  { return d.Depth }

func NewDataChunk(bitDepth uint16, subchunk *SubChunk) DataChunk {
	if subchunk != nil && string(subchunk.Subchunk2ID[:]) == junkSubchunkIDString {
		bitDepth = 0
	}

	switch bitDepth {
	case 0:
		if subchunk == nil {
			subchunk = NewJunkSubChunk()
		}
		return &DataChunkJunk{
			SubChunk: subchunk,
			Depth:    0,
		}
	case bitDepth8:
		if subchunk == nil {
			subchunk = NewDataSubChunk()
		}
		return &DataChunk8bit{
			SubChunk: subchunk,
			Depth:    bitDepth8,
		}
	case bitDepth16:
		if subchunk == nil {
			subchunk = NewDataSubChunk()
		}
		return &DataChunk16bit{
			SubChunk: subchunk,
			Depth:    bitDepth16,
		}
	case bitDepth24:
		if subchunk == nil {
			subchunk = NewDataSubChunk()
		}
		return &DataChunk24bit{
			SubChunk: subchunk,
			Depth:    bitDepth24,
		}
	case bitDepth32:
		if subchunk == nil {
			subchunk = NewDataSubChunk()
		}
		return &DataChunk32bit{
			SubChunk: subchunk,
			Depth:    bitDepth32,
		}
	default:
		return nil
	}
}
