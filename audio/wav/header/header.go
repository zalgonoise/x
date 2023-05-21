package header

import (
	"encoding/binary"
)

const Size = 36

var (
	defaultChunkID     = [4]byte{82, 73, 70, 70}
	defaultFormat      = [4]byte{87, 65, 86, 69}
	defaultSubchunk1ID = [4]byte{102, 109, 116, 32}
)

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

// New creates a Header from the input sampleRate, bitDepth, numChannels and format.
//
// This call also validates the generated header in the returned error
func New(sampleRate uint32, bitDepth, numChannels, format uint16) (*Header, error) {
	h := &Header{
		ChunkID:       defaultChunkID,
		ChunkSize:     0,
		Format:        defaultFormat,
		Subchunk1ID:   defaultSubchunk1ID,
		Subchunk1Size: 16,
		AudioFormat:   format,
		NumChannels:   numChannels,
		SampleRate:    sampleRate,
		ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / 8,
		BlockAlign:    bitDepth * numChannels / 8,
		BitsPerSample: bitDepth,
	}

	return h, Validate(h)
}

// Write implements the io.Writer interface
//
// It consumes the byte slice `buf` as a Wav Header, returning an error
// if the input data cannot be parsed, or if the resulting header is invalid
func (h *Header) Write(buf []byte) (n int, err error) {
	if len(buf) < Size {
		return 0, ErrShortDataBuffer
	}

	h.ChunkID = [4]byte(buf[:4])
	h.ChunkSize = binary.LittleEndian.Uint32(buf[4:8])
	h.Format = [4]byte(buf[8:12])
	h.Subchunk1ID = [4]byte(buf[12:16])
	h.Subchunk1Size = binary.LittleEndian.Uint32(buf[16:20])
	h.AudioFormat = binary.LittleEndian.Uint16(buf[20:22])
	h.NumChannels = binary.LittleEndian.Uint16(buf[22:24])
	h.SampleRate = binary.LittleEndian.Uint32(buf[24:28])
	h.ByteRate = binary.LittleEndian.Uint32(buf[28:32])
	h.BlockAlign = binary.LittleEndian.Uint16(buf[32:34])
	h.BitsPerSample = binary.LittleEndian.Uint16(buf[34:36])

	return len(buf), Validate(h)
}

// Read implements the io.Reader interface
//
// It reads the Header into the byte slice `buf` in Little Endian byte order,
// returning the number of bytes written and an error if raised
func (h *Header) Read(buf []byte) (n int, err error) {
	if len(buf) < Size {
		return 0, ErrShortDataBuffer
	}

	copy(buf[:4], h.ChunkID[:])
	binary.LittleEndian.PutUint32(buf[4:8], h.ChunkSize)
	copy(buf[8:12], h.Format[:])
	copy(buf[12:16], h.Subchunk1ID[:])
	binary.LittleEndian.PutUint32(buf[16:20], h.Subchunk1Size)
	binary.LittleEndian.PutUint16(buf[20:22], h.AudioFormat)
	binary.LittleEndian.PutUint16(buf[22:24], h.NumChannels)
	binary.LittleEndian.PutUint32(buf[24:28], h.SampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], h.ByteRate)
	binary.LittleEndian.PutUint16(buf[32:34], h.BlockAlign)
	binary.LittleEndian.PutUint16(buf[34:36], h.BitsPerSample)

	return Size, nil
}

// Bytes casts a Header as a slice of bytes, by binary-encoding the
// object with a little-endian (LE) byte order
func (h *Header) Bytes() []byte {
	buf := make([]byte, Size)

	if _, err := h.Read(buf); err != nil {
		return nil
	}

	return buf
}
