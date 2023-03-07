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

// Chunk describes the behavior that a data chunk exposes, which involve
// reading and writing PCM audio buffers from / to bytes. Additionally, it
// provides helper methods to fetch the chunk header, the bit depth, to reset it
// and also to retrieve the PCM buffer as a slice of int
type Chunk interface {
	// Parse will consume the input byte slice `buf`, to extract the PCM audio buffer
	// from raw bytes
	Parse(buf []byte)
	// Generate will return a slice of bytes with the encoded PCM buffer
	Generate() []byte
	// Header returns the ChunkHeader of the Chunk
	Header() *ChunkHeader
	// BitDepth returns the bit depth of the Chunk
	BitDepth() uint16
	// Reset clears the data stored in the Chunk
	Reset()
	// Value returns the PCM audio buffer from the Chunk, as a slice of int
	Value() []int
}

// ChunkHeader describes the (raw) structure of a WAV file subchunk, which usually
// contains a "data" or "junk" ID, and the length of the data as its size
type ChunkHeader struct {
	Subchunk2ID   [4]byte // 37-40
	Subchunk2Size uint32  // 41-44
}

// HeaderFrom reads the ChunkHeader from the input byte slice `buf`, returning it and
// an error in case the data is invalid
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
