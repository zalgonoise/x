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
	dataChunkBaseLen     = 512
)

var (
	defaultSubchunk2ID = [4]byte([]byte(dataSubchunkIDString))
	junkSubchunk2ID    = [4]byte([]byte(junkSubchunkIDString))
)

// ChunkHeader describes the (raw) structure of a WAV file subchunk, which usually
// contains a "data" or "junk" ID, and the length of the data as its size
type ChunkHeader struct {
	Subchunk2ID   [4]byte // 37-40 || 1-4
	Subchunk2Size uint32  // 41-44 || 5-8
}

// HeaderFrom reads the ChunkHeader from the input byte slice `buf`, returning it and
// an error in case the data is invalid
func HeaderFrom(buf []byte) (*ChunkHeader, error) {
	var size = buf[4:8:8]
	chunk := &ChunkHeader{
		Subchunk2ID:   [4]byte(buf[:4]),
		Subchunk2Size: binary.LittleEndian.Uint32(size[:]),
	}

	switch string(chunk.Subchunk2ID[:]) {
	case junkSubchunkIDString, dataSubchunkIDString:
		return chunk, nil
	default:
		return nil, ErrInvalidSubChunkHeader
	}
}

// Bytes casts the ChunkHeader `s` as a slice of bytes
func (s *ChunkHeader) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, s)
	return buf.Bytes()
}

// NewDataHeader creates a new ChunkHeader tagged with a "data" ID
func NewDataHeader() *ChunkHeader {
	return &ChunkHeader{Subchunk2ID: defaultSubchunk2ID}
}

// NewJunkHeader creates a new ChunkHeader tagged with a "junk" ID
func NewJunkHeader() *ChunkHeader {
	return &ChunkHeader{Subchunk2ID: junkSubchunk2ID}
}
