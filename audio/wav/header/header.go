package header

import (
	"bytes"
	"encoding/binary"
)

const headerSize = 36

// Header describes the header of a WAV file (or buffer).
//
// The structure is defined as seen in the WAV file format, and can
// be quickly encoded / decoded into binary format as-is
//
// Reference: http://soundfile.sapp.org/doc/WaveFormat/
type Header struct {
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

// From extracts a WAV header from an input chunk of bytes; returning a
// pointer to a Header, and an error if the data is invalid
func From(buf []byte) (h *Header, err error) {
	h = new(Header)

	if _, err = h.Write(buf); err != nil {
		return nil, err
	}

	return h, nil
}

// Write implements the io.Writer interface
//
// It consumes the byte slice `buf` as a Wav Header, returning an error
// if the input data cannot be parsed, or if the resulting header is invalid
func (h *Header) Write(buf []byte) (n int, err error) {
	if err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, h); err != nil {
		return 0, err
	}

	return len(buf), Validate(h)
}

// Read implements the io.Reader interface
//
// It reads the Header into the byte slice `buf` in Little Endian byte order,
// returning the number of bytes written and an error if raised
func (h *Header) Read(buf []byte) (n int, err error) {
	b := bytes.NewBuffer(buf)
	err = binary.Write(b, binary.LittleEndian, h)

	if err != nil {
		return 0, err
	}

	n = len(buf)
	if n < headerSize {
		return n, nil
	}

	return headerSize, nil
}

// Bytes casts a Header as a slice of bytes, by binary-encoding the
// object with a little-endian (LE) byte order
func (h *Header) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, headerLen))
	_ = binary.Write(buf, binary.LittleEndian, h)
	return buf.Bytes()
}
