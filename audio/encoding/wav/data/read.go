package data

import (
	"bytes"
	"errors"
	"io"
	"math"
)

const (
	// defaultReadSize is the read buffer size used when a chunk has no better hint.
	defaultReadSize = 4096
	// maxEmptyReads bounds consecutive (0, nil) reads before giving up, like bufio does.
	maxEmptyReads = 100
)

// ErrNilReader is returned when a nil io.Reader is passed to a ReadFrom method.
var ErrNilReader = errors.New("audio/wav/data: nil reader")

// declaredSize returns the number of bytes a chunk declares, or 0 when the size is unknown: 0 and 0xFFFFFFFF are
// used by live streams to mean "until the end of the stream".
func declaredSize(h *Header) int64 {
	if h == nil || h.Subchunk2Size == 0 || h.Subchunk2Size == math.MaxUint32 {
		return 0
	}

	return int64(h.Subchunk2Size)
}

// chunkSize returns n as a chunk size field, saturated at math.MaxUint32: a RIFF size field cannot describe more.
func chunkSize(n int) uint32 {
	if n < 0 {
		return 0
	}

	return uint32(min(uint64(n), math.MaxUint32))
}

// readFrames reads r until io.EOF, or until limit bytes when limit > 0, calling fn with buffers whose length is always a
// multiple of align (the frame size in bytes). It honours the io.Reader contract:
//   - short reads are accumulated, and a partial frame is carried over to the next read instead of being parsed;
//   - bytes returned together with an error (including io.EOF) are processed before the error is handled;
//   - a trailing partial frame at the end of the input is dropped, since it cannot be decoded.
//
// The buffer passed to fn is reused between calls: fn must not retain it. readSize is the preferred read size; it is
// rounded down to a multiple of align (and never below one frame). Reaching the end of the input returns a nil error.
func readFrames(r io.Reader, align, readSize int, limit int64, fn func([]byte) error) (n int64, err error) {
	if r == nil {
		return 0, ErrNilReader
	}

	if align <= 0 {
		align = 1
	}

	if readSize <= 0 {
		readSize = defaultReadSize
	}

	readSize = max(readSize-readSize%align, align)

	if limit > 0 {
		r = io.LimitReader(r, limit)
	}

	var (
		buf     = make([]byte, readSize)
		carry   int
		empties int
	)

	for {
		m, readErr := r.Read(buf[carry:])

		switch {
		case m < 0:
			return n, io.ErrNoProgress
		case m == 0 && readErr == nil:
			empties++

			if empties >= maxEmptyReads {
				return n, io.ErrNoProgress
			}

			continue
		}

		empties = 0
		n += int64(m)
		total := carry + m

		if whole := total - total%align; whole > 0 {
			if fnErr := fn(buf[:whole]); fnErr != nil {
				return n, fnErr
			}

			carry = copy(buf, buf[whole:total])
		} else {
			carry = total
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return n, nil
			}

			return n, readErr
		}
	}
}

// skipPad consumes the RIFF pad byte that follows a chunk with an odd declared size. A missing pad byte at the end of
// the input is not an error.
func skipPad(r io.Reader, h *Header) (int64, error) {
	if size := declaredSize(h); size&1 == 0 {
		return 0, nil
	}

	var pad [1]byte

	n, err := io.ReadFull(r, pad[:])
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}

	return int64(n), err
}

// readBounded reads exactly the declared size of a chunk into dst (growing it as data arrives, so a lying header cannot
// force a large allocation), then consumes the pad byte. An unknown size reads until io.EOF.
func readBounded(r io.Reader, h *Header, dst *bytes.Buffer) (n int64, err error) {
	if r == nil {
		return 0, ErrNilReader
	}

	size := declaredSize(h)
	if size == 0 {
		return dst.ReadFrom(r)
	}

	n, err = io.CopyN(dst, r, size)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}

		return n, err
	}

	pad, err := skipPad(r, h)

	return n + pad, err
}
