package data

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	Size = 8

	DataIDString = "data"
	JunkIDString = "junk"
)

type ID [4]byte

var (
	//nolint:gochecknoglobals // is served for comparisons and assignments, as a byte array for its string constant
	DataID ID = [4]byte([]byte(DataIDString))
	//nolint:gochecknoglobals // is served for comparisons and assignments, as a byte array for its string constant
	JunkID ID = [4]byte([]byte(JunkIDString))
)

// Header describes the (raw) structure of a WAV file subchunk, which usually
// contains a "data" or "junk" ID, and the length of the data as its size.
//
// Reference: http://soundfile.sapp.org/doc/WaveFormat/
type Header struct {
	Subchunk2ID   [4]byte // 37-40 || 1-4 big endian (4 bytes)
	Subchunk2Size uint32  // 41-44 || 5-8 little endian (4 bytes)
}

// From extracts a subchunk Header from an input slice of bytes; returning a
// pointer to a Header, and an error if the data is invalid.
func From(buf []byte) (h *Header, err error) {
	h = new(Header)

	if _, err = h.Write(buf); err != nil {
		return nil, err
	}

	return h, nil
}

// Write implements the io.Writer interface.
//
// It consumes the byte slice `buf` as a subchunk Header, returning an error
// if the input data cannot be parsed, or if the resulting header is invalid.
func (h *Header) Write(buf []byte) (n int, err error) {
	if len(buf) < Size {
		return 0, ErrShortBuffer
	}

	h.Subchunk2ID = [4]byte(buf[:4])
	h.Subchunk2Size = binary.LittleEndian.Uint32(buf[4:8:8])

	return Size, Validate(h)
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
		return n, ErrShortBuffer
	}

	m, err = h.Write(buf)

	return int64(m), err
}

// Read implements the io.Reader interface.
//
// It reads the Header into the byte slice `buf`, returning the number of bytes written and an error if raised.
func (h *Header) Read(buf []byte) (n int, err error) {
	if len(buf) < Size {
		return 0, ErrShortBuffer
	}

	copy(buf[:4], h.Subchunk2ID[:])
	binary.LittleEndian.PutUint32(buf[4:8], h.Subchunk2Size)

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

// New creates a Header based on the input subchunk ID.
func New(id ID) *Header {
	switch id {
	case DataID, JunkID:
		return &Header{Subchunk2ID: id}
	default:
		return nil
	}
}

// NewDataHeader creates a new ChunkHeader tagged with a "data" ID.
func NewDataHeader() *Header {
	return &Header{Subchunk2ID: DataID}
}

// NewJunkHeader creates a new ChunkHeader tagged with a "junk" ID.
func NewJunkHeader() *Header {
	return &Header{Subchunk2ID: JunkID}
}
