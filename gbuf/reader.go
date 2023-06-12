package gbuf

import (
	"errors"
	"io"

	"github.com/zalgonoise/gio"
)

// A Reader implements the gio.Reader, gio.ReaderAt, gio.WriterTo, gio.Seeker,
// and gio.Scanner interfaces by reading from
// a T item slice.
// Unlike a Buffer, a Reader is read-only and supports seeking.
// The zero value for Reader operates like a Reader of an empty slice.
type Reader[T any] struct {
	s        []T
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

// Len returns the number of T items of the unread portion of the
// slice.
func (r *Reader[T]) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying T item slice.
// Size is the number of T items available for reading via ReadAt.
// The result is unaffected by any method calls except Reset.
func (r *Reader[T]) Size() int64 { return int64(len(r.s)) }

// Read implements the io.Reader interface.
func (r *Reader[T]) Read(b []T) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

// ReadAt implements the io.ReaderAt interface.
func (r *Reader[T]) ReadAt(b []T, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("gbuf.Reader.ReadAt: negative offset")
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

// ReadItem implements the gio.ItemReader interface.
func (r *Reader[T]) ReadItem() (T, error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		var zero T
		return zero, io.EOF
	}
	b := r.s[r.i]
	r.i++
	return b, nil
}

// UnreadItem complements ReadItem in implementing the gio.ItemScanner interface.
func (r *Reader[T]) UnreadItem() error {
	if r.i <= 0 {
		return errors.New("gbuf.Reader.UnreadT item: at beginning of slice")
	}
	r.prevRune = -1
	r.i--
	return nil
}

// Seek implements the gio.Seeker interface.
func (r *Reader[T]) Seek(offset int64, whence int) (int64, error) {
	r.prevRune = -1
	var abs int64
	switch whence {
	case gio.SeekStart:
		abs = offset
	case gio.SeekCurrent:
		abs = r.i + offset
	case gio.SeekEnd:
		abs = int64(len(r.s)) + offset
	default:
		return 0, errors.New("gbuf.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("gbuf.Reader.Seek: negative position")
	}
	r.i = abs
	return abs, nil
}

// WriteTo implements the gio.WriterTo interface.
func (r *Reader[T]) WriteTo(w gio.Writer[T]) (n int64, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	b := r.s[r.i:]
	m, err := w.Write(b)
	if m > len(b) {
		panic("gbuf.Reader.WriteTo: invalid Write count")
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(b) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

// Reset resets the Reader to be reading from b.
func (r *Reader[T]) Reset(b []T) { *r = Reader[T]{b, 0, -1} }

// NewReader returns a new Reader reading from b.
func NewReader[T any](b []T) *Reader[T] { return &Reader[T]{b, 0, -1} }
