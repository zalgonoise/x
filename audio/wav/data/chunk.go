package data

import (
	"bytes"
	"encoding/binary"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrInvalidSubChunkHeader err = "audio/wav: invalid subchunk header metadata"

	dataSubchunkIDString = "data"
	junkSubchunkIDString = "junk"
)

var (
	defaultSubchunk2ID = [4]byte{100, 97, 116, 97}
	junkSubchunk2ID    = [4]byte{106, 117, 110, 107}
)

type Chunk interface {
	Parse(buf []byte, offset int)
	Generate() []byte
	Header() *ChunkHeader
	BitDepth() uint16
}

type ChunkHeader struct {
	Subchunk2ID   [4]byte // 37-40
	Subchunk2Size uint32  // 41-44
}

func HeaderFrom(buf []byte) (*ChunkHeader, error) {
	r := bytes.NewReader(buf)
	var chunk = new(ChunkHeader)
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

func (s *ChunkHeader) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, s)
	return buf.Bytes()
}

func NewDataHeader() *ChunkHeader {
	return &ChunkHeader{Subchunk2ID: defaultSubchunk2ID}
}

func NewJunkHeader() *ChunkHeader {
	return &ChunkHeader{Subchunk2ID: junkSubchunk2ID}
}
