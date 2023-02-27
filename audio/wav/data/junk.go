package data

type ChunkJunk struct {
	*ChunkHeader
	Data  []byte
	Depth uint16
}

func (d *ChunkJunk) Parse(buf []byte, offset int) {
	if d.Data == nil {
		d.Data = buf
		if d.Subchunk2Size == 0 {
			d.Subchunk2Size = uint32(len(buf))
		}
		return
	}
	d.Data = append(d.Data, buf...)
}

func (d *ChunkJunk) Generate() []byte {
	return d.Data
}

func (d *ChunkJunk) Header() *ChunkHeader { return d.ChunkHeader }
func (d *ChunkJunk) BitDepth() uint16     { return d.Depth }
