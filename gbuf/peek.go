package gbuf

import (
	"fmt"
	"io"
)

// Peek reads from the Buffer `b`, however it does not advance the buffer's offset after the items are read
func Peek[T any](p []T, b *Buffer[T]) (n int, err error) {
	if b == nil {
		return 0, nil
	}

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
func PeekFrom[T any](idx int, p []T, b *Buffer[T]) (n int, err error) {
	if b == nil {
		return 0, nil
	}

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
func PeekRange[T any](from, to int, p []T, b *Buffer[T]) (n int, err error) {
	if b == nil {
		return 0, nil
	}

	if from == to {
		return 0, nil
	}

	var (
		invert bool
		ln     = len(p)
	)

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
