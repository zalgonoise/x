package gbuf

// Simple generic type buffer for marshaling data.

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/zalgonoise/gio"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

// A Buffer is a variable-sized buffer of T items with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer[T any] struct {
	buf       []T         // contents are the T items buf[off : len(buf)]
	off       int         // read at &buf[off], write at &buf[len(buf)]
	isReading atomic.Bool // last operation type, so that Unread* can work correctly.
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("gbuf.Buffer: too large")
var errNegativeRead = errors.New("gbuf.Buffer: reader returned negative count from Read")

const maxInt = int(^uint(0) >> 1)

// Value returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (b *Buffer[T]) Value() []T { return b.buf[b.off:] }

// empty reports whether the unread portion of the buffer is empty.
func (b *Buffer[T]) empty() bool { return len(b.buf) <= b.off }

// Len returns the number of T items of the unread portion of the buffer;
// b.Len() == len(b.Value()).
func (b *Buffer[T]) Len() int { return len(b.buf) - b.off }

// Cap returns the capacity of the buffer's underlying T item slice, that is, the
// total space allocated for the buffer's data.
func (b *Buffer[T]) Cap() int { return cap(b.buf) }

// Truncate discards all but the first n unread T items from the buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than the length of the buffer.
func (b *Buffer[T]) Truncate(n int) {
	if n == 0 {
		b.Reset()
		return
	}

	b.isReading.Store(false)

	if n < 0 || n > b.Len() {
		panic("gbuf.Buffer: truncation out of range")
	}

	b.buf = b.buf[:b.off+n]
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *Buffer[T]) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.isReading.Store(false)
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where T items should be written and whether it succeeded.
func (b *Buffer[T]) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more T items.
// It returns the index where T items should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer[T]) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}

	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]T, n, smallBufferSize)
		return 0
	}

	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Add b.off to account for b.buf[:b.off] being sliced off the front.
		b.buf = growSlice(b.buf[b.off:], b.off+n)
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n T items. After Grow(n), at least n T items can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer[T]) Grow(n int) {
	if n < 0 {
		panic("gbuf.Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer[T]) Write(p []T) (n int, err error) {
	b.isReading.Store(false)
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

// MinRead is the minimum slice size passed to a Read call by
// Buffer.ReadFrom. As long as the Buffer has at least MinRead T items beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of T items read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *Buffer[T]) ReadFrom(r gio.Reader[T]) (n int64, err error) {
	b.isReading.Store(false)
	for {
		i := b.grow(MinRead)
		b.buf = b.buf[:i]
		m, e := r.Read(b.buf[i:cap(b.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}

		b.buf = b.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}
	}
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice[T any](b []T, n int) []T {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]T item, n)...)
	// This avoids unnecessary zero-ing of the first len(b) T items of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]T(nil), make([]T, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of T items written; it always fits into an
// int, but it is int64 to match the gio.WriterTo interface. Any error
// encountered during the write operation is also returned.
func (b *Buffer[T]) WriteTo(w gio.Writer[T]) (n int64, err error) {
	b.isReading.Store(false)
	if nItems := b.Len(); nItems > 0 {
		m, e := w.Write(b.buf[b.off:])
		if m > nItems {
			panic("gbuf.Buffer.WriteTo: invalid Write count")
		}
		b.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		// all T items should have been written, by definition of
		// Write method in gio.Writer
		if m != nItems {
			return n, io.ErrShortWrite
		}
	}
	// Buffer is now empty; reset.
	b.Reset()
	return n, nil
}

// WriteItem appends the T `item` to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match gio.Writer's
// WriteItem. If the buffer becomes too large, WriteItem will panic with
// ErrTooLarge.
func (b *Buffer[T]) WriteItem(item T) error {
	b.isReading.Store(false)
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = item
	return nil
}

// Read reads the next len(p) T items from the buffer or until the buffer
// is drained. The return value n is the number of T items read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer[T]) Read(p []T) (n int, err error) {
	b.isReading.Store(false)
	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	if n > 0 {
		b.isReading.Store(true)
	}
	return n, nil
}

// Next returns a slice containing the next n T items from the buffer,
// advancing the buffer as if the T items had been returned by Read.
// If there are fewer than n T items in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (b *Buffer[T]) Next(n int) []T {
	b.isReading.Store(false)
	m := b.Len()
	if n > m {
		n = m
	}
	data := b.buf[b.off : b.off+n]
	b.off += n
	if n > 0 {
		b.isReading.Store(true)
	}
	return data
}

// ReadItem reads and returns the next T item from the buffer.
// If no T item is available, it returns error io.EOF.
func (b *Buffer[T]) ReadItem() (T, error) {
	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		var zero T
		return zero, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	b.isReading.Store(true)
	return c, nil
}

var ErrUnreadItem = errors.New("gbuf.Buffer: UnreadItem: previous operation was not a successful read")

// UnreadItem unreads the last T item returned by the most recent successful
// read operation that read at least one T item. If a write operation has happened since
// the last read, if the last read returned an error, or if the read operation reads zero
// T items, UnreadItem returns an error.
func (b *Buffer[T]) UnreadItem() error {
	if !b.isReading.Load() {
		return ErrUnreadItem
	}
	b.isReading.Store(false)
	if b.off > 0 {
		b.off--
	}
	return nil
}

// ReadItems reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadT items encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadT items returns err != nil if and only if the returned data does not end in
// delim.
func (b *Buffer[T]) ReadItems(delim func(T) bool) (line []T, err error) {
	slice, err := b.readSlice(delim)
	// return a copy of slice. The buffer's backing array may
	// be overwritten by later calls.
	line = append(line, slice...)
	return line, err
}

// readSlice is like ReadItems but returns a reference to internal buffer data.
func (b *Buffer[T]) readSlice(delim func(T) bool) (line []T, err error) {
	// init end as not found
	var end = -1

	// iterate through remaining buffer trying to find a matching delimiter
	for i := b.off; i < len(b.buf); i++ {
		if delim(b.buf[i]) {
			// set end to index + 1, as slice ranges require it
			end = i + 1
			break
		}
	}

	// if there are no matches, get up to the remainder of the existing buffer
	if end < 0 {
		end = len(b.buf)
		err = io.EOF
	}

	line = b.buf[b.off:end]
	b.off = end
	b.isReading.Store(true)

	return line, err
}

// NewBuffer creates and initializes a new Buffer using buf as its
// initial contents. The new Buffer takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer[T]) (or just declaring a *Buffer[T] variable) is
// sufficient to initialize a Buffer of type T.
func NewBuffer[T any](buf []T) *Buffer[T] {
	return &Buffer[T]{
		buf: buf,
	}
}
