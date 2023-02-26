package wav

import (
	"bytes"
	"encoding/binary"
)

const ErrInvalidSubChunkHeader err = "invalid subchunk header metadata"

type SubChunk struct {
	Subchunk2ID   [4]byte // 37-40
	Subchunk2Size uint32  // 41-44
}

func SubChunkFrom(buf []byte) (*SubChunk, error) {
	r := bytes.NewReader(buf)
	var chunk = new(SubChunk)
	err := binary.Read(r, binary.LittleEndian, chunk)
	if err != nil {
		return nil, err
	}
	switch string(chunk.Subchunk2ID[:]) {
	case junkSubchunkIDString, dataSubchunkIDString:
		return chunk, nil
	default:
		return nil, ErrInvalidSubChunkHeader
	}
}

func (s *SubChunk) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, s)
	return buf.Bytes()
}

func NewDataSubChunk() *SubChunk {
	return &SubChunk{Subchunk2ID: defaultSubchunk2ID}
}

func NewJunkSubChunk() *SubChunk {
	return &SubChunk{Subchunk2ID: junkSubchunk2ID}
}
