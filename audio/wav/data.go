package wav

import (
	"encoding/binary"
)

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
		d.Data = make([]byte, len(buf))
		copy(d.Data, buf)
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
		d.Data = make([]int8, len(buf))
		for i := 0; i < len(buf); i++ {
			d.Data[i] = int8(buf[i])
		}
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	new := make([]int8, len(buf))
	for i := 0; i < len(buf); i++ {
		new[i] = int8(buf[i])
	}
	d.Data = append(d.Data, new...)
}

func (d *DataChunk8bit) Generate() []byte {
	data := make([]byte, len(d.Data))
	for i := 0; i < len(d.Data); i++ {
		data[i] = byte(d.Data[i])
	}
	return data
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
		d.Data = make([]int16, len(buf)/2)
		for i, j := 0, 0; i+1 < len(buf); i, j = i+2, j+1 {
			d.Data[j] = int16(binary.LittleEndian.Uint16(buf[i : i+2]))
		}
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	new := make([]int16, len(buf)/2)
	for i, j := 0, 0; i+1 < len(buf); i, j = i+2, j+1 {
		new[j] = int16(binary.LittleEndian.Uint16(buf[i : i+2]))
	}
	d.Data = append(d.Data, new...)
}

func (d *DataChunk16bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*2)
	for i := 0; i < len(d.Data); i++ {
		data = binary.LittleEndian.AppendUint16(data, uint16(d.Data[i]))
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
		d.Data = make([]int32, len(buf)/3)
		for i, j := 0, 0; i+2 < len(buf); i, j = i+3, j+1 {
			d.Data[j] = int32(decode24BitLE(buf[i : i+3]))
		}
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	new := make([]int32, len(buf)/3)
	for i, j := 0, 0; i+2 < len(buf); i, j = i+3, j+1 {
		new[j] = int32(decode24BitLE(buf[i : i+3]))
	}

	d.Data = append(d.Data, new...)
}

func (d *DataChunk24bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*3)
	for i := 0; i < len(d.Data); i++ {
		data = encode24BitLE(data, int32(d.Data[i]))
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
		d.Data = make([]int32, len(buf)/4)
		for i, j := 0, 0; i+3 < len(buf); i, j = i+4, j+1 {
			d.Data[j] = int32(binary.LittleEndian.Uint32(buf[i : i+4]))
		}
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	new := make([]int32, len(buf)/4)
	for i, j := 0, 0; i+3 < len(buf); i, j = i+4, j+1 {
		new[j] = int32(binary.LittleEndian.Uint32(buf[i : i+4]))
	}
	d.Data = append(d.Data, new...)
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
