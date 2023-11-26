package wav

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	Size                 = 36
	byteSize             = 8
	chunkIDEnd           = 4
	formatOffset         = 8
	subChunkIDEnd        = 16
	defaultSubChunk1Size = 16
)

//nolint:gochecknoglobals // referenced directly for assignments and comparisons
var (
	defaultChunkID      = [4]byte{82, 73, 70, 70}
	defaultFormat       = [4]byte{87, 65, 86, 69}
	defaultSubchunk1ID  = [4]byte{102, 109, 116, 32}
	formatAndSubchunkID = []byte{87, 65, 86, 69, 102, 109, 116, 32}
)

// Header describes the header of a WAV file (or buffer).
//
// The structure is defined as seen in the WAV file format, and can
// be quickly encoded / decoded into binary format as-is.
//
// Reference: http://soundfile.sapp.org/doc/WaveFormat/
type Header struct {
	ChunkID       [4]byte // 1-4 big endian (4 bytes)
	ChunkSize     uint32  // 5-8 little endian (4 bytes)
	Format        [4]byte // 9-12 big endian (4 bytes)
	Subchunk1ID   [4]byte // 13-16 big endian (4 bytes)
	Subchunk1Size uint32  // 17-20 little endian (4 bytes)
	AudioFormat   uint16  // 21-22 little endian (2 bytes)
	NumChannels   uint16  // 23-24 little endian (2 bytes)
	SampleRate    uint32  // 25-28 little endian (4 bytes)
	ByteRate      uint32  // 29-32 little endian (4 bytes)
	BlockAlign    uint16  // 33-34 little endian (2 bytes)
	BitsPerSample uint16  // 35-36 little endian (2 bytes)
}

// HeaderFrom extracts a WAV Header from an input slice of bytes; returning a
// pointer to a Header, and an error if the data is invalid.
func HeaderFrom(buf []byte) (h *Header, err error) {
	h = new(Header)

	if _, err = h.Write(buf); err != nil {
		return nil, err
	}

	return h, nil
}

// NewHeader creates a Header from the input sampleRate, bitDepth, numChannels and format.
//
// This call also validates the generated header in the returned error.
func NewHeader(sampleRate uint32, bitDepth, numChannels, format uint16) (*Header, error) {
	h := &Header{
		ChunkID:       defaultChunkID,
		ChunkSize:     0,
		Format:        defaultFormat,
		Subchunk1ID:   defaultSubchunk1ID,
		Subchunk1Size: defaultSubChunk1Size,
		AudioFormat:   format,
		NumChannels:   numChannels,
		SampleRate:    sampleRate,
		ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / byteSize,
		BlockAlign:    bitDepth * numChannels / byteSize,
		BitsPerSample: bitDepth,
	}

	return h, ValidateHeader(h)
}

// Write implements the io.Writer interface.
//
// It consumes the byte slice `buf` as a Wav Header, returning an error
// if the input data cannot be parsed, or if the resulting header is invalid.
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

	return Size, ValidateHeader(h)
}

// ReadFrom implements the io.ReaderFrom interface.
//
// It consumes the byte slice `buf` as a Wav Header from an io.Reader, returning an error
// if the input data cannot be parsed, or if the resulting header is invalid.
func (h *Header) ReadFrom(r io.Reader) (n int64, err error) {
	if r == nil {
		return 0, nil
	}

	// required as it cannot be just cast into the data type
	buf := make([]byte, Size)

	m, err := r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if m != Size {
		return n, ErrShortDataBuffer
	}

	m, err = h.Write(buf)

	return int64(m), err
}

// Read implements the io.Reader interface.
//
// It reads the Header into the byte slice `buf`, returning the number of bytes written and an error if raised.
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
// object.
func (h *Header) Bytes() []byte {
	buf := make([]byte, Size)

	if _, err := h.Read(buf); err != nil {
		return nil
	}

	return buf
}

// GetSampleRate returns the SampleRate value from the Header, to satisfy a
// common header interface among different audio encodings.
func (h *Header) GetSampleRate() int {
	if h == nil {
		return 0
	}

	return int(h.SampleRate)
}
