package wav

import (
	"bytes"
	"encoding/binary"
)

type WavHeader struct {
	ChunkID       [4]byte // 1-4
	ChunkSize     uint32  // 5-8
	Format        [4]byte // 9-12
	Subchunk1ID   [4]byte // 13-16
	Subchunk1Size uint32  // 17-20
	AudioFormat   uint16  // 21-22
	NumChannels   uint16  // 23-24
	SampleRate    uint32  // 25-28
	ByteRate      uint32  // 29-32
	BlockAlign    uint16  // 33-34
	BitsPerSample uint16  // 35-36
}

func HeaderFrom(buf []byte) (*WavHeader, error) {
	r := bytes.NewReader(buf)
	var header = new(WavHeader)
	err := binary.Read(r, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}
	if string(header.ChunkID[:]) != string(defaultChunkID[:]) ||
		string(header.Format[:]) != string(defaultFormat[:]) {
		return nil, ErrInvalidHeader
	}
	return header, nil
}

func (h *WavHeader) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, headerLen))
	_ = binary.Write(buf, binary.LittleEndian, h)
	return buf.Bytes()
}
