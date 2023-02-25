package wav

import "encoding/binary"

type DataChunk interface {
	Parse(buf []byte)
	Generate() []byte
	Header() SubChunk
}

type DataChunkJunk struct {
	SubChunk
	Data []byte
}

func (d *DataChunkJunk) Parse(buf []byte) {
	d.Data = buf
	d.Subchunk2Size = uint32(len(buf))
}

func (d *DataChunkJunk) Generate() []byte {
	return d.Data
}

func (d *DataChunkJunk) Header() SubChunk {
	if d.Subchunk2Size == 0 {
		d.Subchunk2Size = uint32(len(d.Data))
	}
	return d.SubChunk
}

type DataChunk8bit struct {
	SubChunk
	Data []int8
}

func (d *DataChunk8bit) Parse(buf []byte) {
	d.Data = make([]int8, len(buf))
	for i := 0; i < len(buf); i++ {
		d.Data = append(d.Data, int8(buf[i]))
	}
	d.Subchunk2Size = uint32(len(buf))
}

func (d *DataChunk8bit) Generate() []byte {
	data := make([]byte, len(d.Data))
	for i := 0; i < len(d.Data); i++ {
		data[i] = byte(d.Data[i])
	}
	return data
}

func (d *DataChunk8bit) Header() SubChunk {
	if d.Subchunk2Size == 0 {
		d.Subchunk2Size = uint32(len(d.Data))
	}
	return d.SubChunk
}

type DataChunk16bit struct {
	SubChunk
	Data []int16
}

func (d *DataChunk16bit) Parse(buf []byte) {
	d.Data = make([]int16, len(buf)/2)
	for i := 0; i+1 < len(buf); i = i + 2 {
		d.Data = append(d.Data, int16(binary.LittleEndian.Uint16(buf[i:i+2])))
	}
	d.Subchunk2Size = uint32(len(buf))
}

func (d *DataChunk16bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*2)
	for i := 0; i < len(d.Data); i++ {
		data = binary.LittleEndian.AppendUint16(data, uint16(d.Data[i]))
	}
	return data
}

func (d *DataChunk16bit) Header() SubChunk {
	if d.Subchunk2Size == 0 {
		d.Subchunk2Size = uint32(len(d.Data) * 2)
	}
	return d.SubChunk
}

type DataChunk24bit struct {
	SubChunk
	Data []int32
}

func (d *DataChunk24bit) Parse(buf []byte) {
	d.Data = make([]int32, len(buf)/3)
	for i := 0; i+2 < len(buf); i = i + 3 {
		d.Data = append(d.Data, int32(decode24BitLE(buf[i:i+3])))
	}
	d.Subchunk2Size = uint32(len(buf))
}

func (d *DataChunk24bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*3)
	for i := 0; i < len(d.Data); i++ {
		data = encode24BitLE(data, int32(d.Data[i]))
	}
	return data
}

func (d *DataChunk24bit) Header() SubChunk {
	if d.Subchunk2Size == 0 {
		d.Subchunk2Size = uint32(len(d.Data) * 3)
	}
	return d.SubChunk
}

type DataChunk32bit struct {
	SubChunk
	Data []int32
}

func (d *DataChunk32bit) Parse(buf []byte) {
	d.Data = make([]int32, len(buf)/4)
	for i := 0; i+3 < len(buf); i = i + 4 {
		d.Data = append(d.Data, int32(binary.LittleEndian.Uint32(buf[i:i+4])))
	}
	d.Subchunk2Size = uint32(len(buf))
}

func (d *DataChunk32bit) Generate() []byte {
	data := make([]byte, 0, len(d.Data)*4)
	for i := 0; i < len(d.Data); i++ {
		data = binary.LittleEndian.AppendUint32(data, uint32(d.Data[i]))
	}
	return data
}

func (d *DataChunk32bit) Header() SubChunk {
	if d.Subchunk2Size == 0 {
		d.Subchunk2Size = uint32(len(d.Data) * 4)
	}
	return d.SubChunk
}

func NewDataChunk(bitDepth uint16) DataChunk {
	switch bitDepth {
	case 0:
		return &DataChunkJunk{
			SubChunk: SubChunk{Subchunk2ID: junkSubchunk2ID},
		}
	case bitDepth8:
		return &DataChunk8bit{
			SubChunk: SubChunk{Subchunk2ID: defaultSubchunk2ID},
		}
	case bitDepth16:
		return &DataChunk16bit{
			SubChunk: SubChunk{Subchunk2ID: defaultSubchunk2ID},
		}
	case bitDepth24:
		return &DataChunk24bit{
			SubChunk: SubChunk{Subchunk2ID: defaultSubchunk2ID},
		}
	case bitDepth32:
		return &DataChunk32bit{
			SubChunk: SubChunk{Subchunk2ID: defaultSubchunk2ID},
		}
	default:
		return nil
	}
}
