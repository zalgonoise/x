package header

import (
	"bytes"
	"encoding/binary"
)

const (
	Size = 8

	DataIDString = "data"
	JunkIDString = "junk"
)

// Header describes the (raw) structure of a WAV file subchunk, which usually
// contains a "data" or "junk" ID, and the length of the data as its size
type Header struct {
	Subchunk2ID   [4]byte // 37-40 || 1-4
	Subchunk2Size uint32  // 41-44 || 5-8
}

func From(buf []byte) (h *Header, err error) {
	h = new(Header)

	if _, err = h.Write(buf); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Header) Write(buf []byte) (n int, err error) {
	var size = buf[4:8:8]
	h.Subchunk2ID = [4]byte(buf[:4])
	h.Subchunk2Size = binary.LittleEndian.Uint32(size[:])

	return Size, Validate(h)
}

func (h *Header) Read(buf []byte) (n int, err error) {
	if len(buf) < Size {
		return 0, ErrShortBuffer
	}

	b := bytes.NewBuffer(buf)
	if err = binary.Write(b, binary.LittleEndian, h); err != nil {
		return 0, err
	}

	return Size, nil
}

func (h *Header) Bytes() []byte {
	buf := make([]byte, 0, 8)

	if _, err := h.Read(buf); err != nil {
		return nil
	}

	return buf
}

// NewData creates a new ChunkHeader tagged with a "data" ID
func NewData() *Header {
	return &Header{Subchunk2ID: [4]byte([]byte(DataIDString))}
}

// NewJunk creates a new ChunkHeader tagged with a "junk" ID
func NewJunk() *Header {
	return &Header{Subchunk2ID: [4]byte([]byte(JunkIDString))}
}
