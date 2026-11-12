package iox

import (
	"errors"
	"fmt"
	"io"
	"math"
)

// Errors returned by SeekBuffer.
var (
	// ErrNegativeOffset is returned for a seek or write before the start of the buffer.
	ErrNegativeOffset = errors.New("iox: negative offset")
	// ErrWhence is returned by Seek for an unknown whence.
	ErrWhence = errors.New("iox: invalid whence")
	// ErrTooLarge is returned for a write that would end beyond the largest int64 offset.
	ErrTooLarge = errors.New("iox: offset too large")
)

// growthFactor is how much a SeekBuffer's capacity at least grows by when it must reallocate.
const growthFactor = 2

// SeekBuffer is an in-memory io.WriteSeeker and io.WriterAt, so encoders that back-patch sizes (riff.Writer,
// wav.Encoder) can build a complete file in memory. Writing past the end grows it, filling any gap with zeros. The
// zero value is an empty buffer.
type SeekBuffer struct {
	b   []byte
	pos int64
}

// Bytes returns the contents. The slice aliases the buffer until the next write.
func (s *SeekBuffer) Bytes() []byte { return s.b }

// Len returns the size of the contents.
func (s *SeekBuffer) Len() int { return len(s.b) }

// Reset empties the buffer, keeping its capacity.
func (s *SeekBuffer) Reset() { s.b, s.pos = s.b[:0], 0 }

// Write writes p at the current position and advances it.
func (s *SeekBuffer) Write(p []byte) (int, error) {
	n, err := s.WriteAt(p, s.pos)
	s.pos += int64(n)

	return n, err
}

// WriteAt writes p at offset off without moving the position.
func (s *SeekBuffer) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegativeOffset, off)
	}

	if off > math.MaxInt64-int64(len(p)) {
		return 0, fmt.Errorf("%w: %d + %d", ErrTooLarge, off, len(p))
	}

	if end := off + int64(len(p)); end > int64(len(s.b)) {
		if end > int64(cap(s.b)) {
			grown := make([]byte, len(s.b), max(end, growthFactor*int64(cap(s.b))))
			copy(grown, s.b)
			s.b = grown
		}

		clear(s.b[len(s.b):end])
		s.b = s.b[:end]
	}

	return copy(s.b[off:], p), nil
}

// Seek implements io.Seeker.
func (s *SeekBuffer) Seek(offset int64, whence int) (int64, error) {
	var base int64

	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = s.pos
	case io.SeekEnd:
		base = int64(len(s.b))
	default:
		return 0, fmt.Errorf("%w: %d", ErrWhence, whence)
	}

	if base+offset < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegativeOffset, base+offset)
	}

	s.pos = base + offset

	return s.pos, nil
}
