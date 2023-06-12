package gbuf

import (
	"errors"
	"fmt"
	"io"
)

// ErrIndexOutOfBounds is caused when a PeekBuffer method call is made involving an index
// that is out of bounds of the underlying slice
var ErrIndexOutOfBounds = errors.New("gbuf.PeekBuffer: index out of bounds")

// A PeekBuffer is a variable-sized buffer of T items with Read and Write methods.
// The zero value for PeekBuffer is an empty buffer ready to use.
//
// It is an extension of the Buffer type, that exposes Peek, PeekFrom and PeekRange methods.
type PeekBuffer[T any] struct {
	*Buffer[T]
}

// Peek is just like Read, however it does not advance the buffer's offset after the items are read
func (b *PeekBuffer[T]) Peek(p []T) (n int, err error) {
	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	return n, nil
}

// PeekFrom is just like Peek, but it reads from the buffer starting at offset `idx`
func (b *PeekBuffer[T]) PeekFrom(idx int, p []T) (n int, err error) {
	if idx < 0 || idx >= len(b.buf) {
		return 0, ErrIndexOutOfBounds
	}

	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[idx:])
	return n, nil
}

// PeekRange is just like Peek, but it reads from the buffer starting at offset `from` until offset `to`
func (b *PeekBuffer[T]) PeekRange(from, to int, p []T) (n int, err error) {
	var (
		invert bool
		ln     = len(p)
	)

	if from == to {
		return 0, nil
	}

	if from < 0 || from >= len(b.buf) {
		return 0, fmt.Errorf("%w: from value: %d", ErrIndexOutOfBounds, from)
	}

	if to < 0 || to >= len(b.buf) {
		return 0, fmt.Errorf("%w: to value: %d", ErrIndexOutOfBounds, to)
	}

	if from > to {
		invert = true
		to, from = from, to
	}

	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		if ln == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}

	n = copy(p, b.buf[to:from])

	if invert {
		for i, j := 0, ln-1; i < ln/2; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
	}

	return n, nil
}

// NewPeekBuffer creates and initializes a new Buffer using buf as its
// initial contents, as a PeekBuffer. The new Buffer takes ownership of buf,
// and the caller should not use buf after this call. NewPeekBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
func NewPeekBuffer[T any](buf []T) *PeekBuffer[T] {
	return &PeekBuffer[T]{
		Buffer: NewBuffer(buf),
	}
}

// AsPeekBuffer converts a Buffer[T] into a PeekBuffer[T]
func AsPeekBuffer[T any](buf *Buffer[T]) *PeekBuffer[T] {
	return &PeekBuffer[T]{
		Buffer: buf,
	}
}
