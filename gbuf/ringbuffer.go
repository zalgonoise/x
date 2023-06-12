package gbuf

import (
	"errors"
	"io"

	"github.com/zalgonoise/gio"
)

const defaultBufferSize = 256

// RingBuffer is a buffer that is connected end-to-end, which allows continuous
// reads and writes provided that the caller is aware of potential loss of read data
// (as elements are overwritten if not read)
type RingBuffer[T any] struct {
	start int
	end   int
	items []T
}

// Write sets the contents of `p` to the buffer, in sequential order,
// looping through the buffer if needed. The return value n is the
// length of p; err is always nil. If the index in the buffer has not
// been yet read, it will be overwritten
func (r *RingBuffer[T]) Write(p []T) (n int, err error) {
	for i := range p {
		r.items[r.end] = p[i]
		r.end = (r.end + 1) % len(r.items)
		if r.end == r.start {
			r.start = (r.start + 1) % len(r.items)
		}
	}
	return len(p), nil
}

// WriteItem writes the T `item` to the buffer in the next position
// The returned error is always nil, but is included to match gio.Writer's
// WriteItem. If the index in the buffer has not been yet read, it will be
// overwritten
func (r *RingBuffer[T]) WriteItem(item T) (err error) {
	r.items[r.end] = item
	r.end = (r.end + 1) % len(r.items)
	if r.end == r.start {
		r.start = (r.start + 1) % len(r.items)
	}
	return nil
}

// Read reads the next len(p) T items from the buffer or until the buffer
// is drained. The return value n is the number of T items read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (r *RingBuffer[T]) Read(p []T) (n int, err error) {
	if r.start == r.end {
		return 0, io.EOF
	}
	if r.start < r.end {
		n = copy(p, r.items[r.start:r.end])
	} else {
		n = copy(p, r.items[r.start-1:])
		if n < len(p) {
			n += copy(p[n:], r.items[:r.end])
		}
	}
	r.start = (r.start + n) % len(r.items)
	return n, nil
}

// Value returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
func (r *RingBuffer[T]) Value() []T {
	var (
		n     int
		items []T
	)
	if r.start == r.end {
		return nil
	}
	if r.start < r.end {
		items = make([]T, r.end-r.start)
		n = copy(items, r.items[r.start:r.end])
	} else {
		items = make([]T, len(r.items))
		n = copy(items, r.items[r.start-1:])
		if n < len(items) {
			n += copy(items[n:], r.items[:r.end])
		}
	}
	r.start = (r.start + n) % len(r.items)
	return items
}

// Len returns the number of T items of the unread portion of the buffer;
// b.Len() == len(b.T items()).
func (r *RingBuffer[T]) Len() int {
	if r.start < r.end {
		return r.end - r.start
	}
	return len(r.items)
}

// Cap returns the length of the buffer's underlying T item slice, that is, the
// total ring buffer's capacity.
func (r *RingBuffer[T]) Cap() int {
	return len(r.items)
}

// Truncate serves as an alias to Reset(); to preserve the ring buffer size
func (r *RingBuffer[T]) Truncate(n int) {
	r.Reset()
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate().
func (r *RingBuffer[T]) Reset() {
	r.start = 0
	r.end = 0
}

// ReadFrom reads data from b until EOF and appends it to the buffer, cycling
// the buffer as needed. Any unready bytes will be overwritten on each cycle.
// The return value n is the number of T items read. Any error except io.EOF
// encountered during the read is also returned.
func (r *RingBuffer[T]) ReadFrom(b gio.Reader[T]) (n int64, err error) {
	for {
		if r.start < r.end {
			num, err := b.Read(r.items[r.start:r.end])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			if err != nil {
				return n, err
			}
		} else {
			num, err := b.Read(r.items[r.start:len(r.items)])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			if err != nil {
				return n, err
			}
			num, err = b.Read(r.items[:r.end])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			if err != nil {
				return n, err
			}
		}
	}
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of T items written; it always fits into an
// int, but it is int64 to match the gio.WriterTo interface. Any error
// encountered during the write is also returned.
func (r *RingBuffer[T]) WriteTo(b gio.Writer[T]) (n int64, err error) {
	for {
		if r.start < r.end {
			num, err := b.Write(r.items[r.start:r.end])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			if err != nil {
				return n, err
			}
		} else {
			num, err := b.Write(r.items[r.start:len(r.items)])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				return n, nil
			}
			if err != nil {
				return n, err
			}
			num, err = b.Write(r.items[:r.end])
			if n < 0 {
				panic(errNegativeRead)
			}
			n += int64(num)
			if errors.Is(err, io.EOF) {
				r.Reset()
				return n, nil
			}
			if err != nil {
				return n, err
			}
		}
	}
}

func (r *RingBuffer[T]) Next(n int) (items []T) {
	if n == 0 {
		return nil
	}
	if n < 0 || n > r.Len() {
		panic("gbuf.RingBuffer: out of range")
	}

	if r.start+n < r.end {
		items = append(items, r.items[r.start:r.start+n]...)
		r.start += n
		return items
	}

	items = append(items, r.items[r.start-1:]...)
	items = append(items, r.items[:n]...)
	r.start = n
	return items
}

func (r *RingBuffer[T]) UnreadItem() error {
	if r.start == r.end {
		return ErrUnreadItem
	}
	r.start = r.start - 1%len(r.items)
	return nil
}

func (r *RingBuffer[T]) ReadItems(delim func(T) bool) (line []T, err error) {
	if r.start == r.end {
		return line, io.EOF
	}
	if r.start < r.end {
		var i int
		for i = r.start; i < r.end; i++ {
			if delim(r.items[i]) {
				break
			}
		}
		line = append(line, r.items[r.start:r.start+i+1]...)
		r.start += i + 1
		return line, nil
	}

	var i int
	var done bool
	for i = r.start; i < len(r.items); i++ {
		if delim(r.items[i]) {
			done = true
			break
		}
	}
	line = append(line, r.items[r.start:r.start+i+1]...)
	if !done {
		for i = 0; i < r.end; i++ {
			if delim(r.items[i]) {
				break
			}
		}
		line = append(line, r.items[r.start:r.start+i+1]...)
	}
	return line, nil
}

// ReadItem reads and returns the next T item from the buffer.
// If no T item is available, it returns error io.EOF.
func (r *RingBuffer[T]) ReadItem() (item T, err error) {
	if r.start == r.end {
		return item, io.EOF
	}
	item = r.items[r.start]
	r.start = (r.start + 1) % len(r.items)
	return item, nil
}

// Seek implements the gio.Seeker interface. All valid whence will point to
// the current cursor's (read) position
func (r *RingBuffer[T]) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case gio.SeekStart, gio.SeekCurrent, gio.SeekEnd:
		if (r.start + int(offset)) < len(r.items) {
			abs = int64(r.start) + offset
		} else {
			abs = offset - int64(len(r.items)-r.start)
		}
	default:
		return 0, errors.New("gbuf.RingBuffer.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("gbuf.RingBuffer.Seek: negative position")
	}
	r.start = int(abs)
	return abs, nil
}

// NewRingBuffer creates a RingBuffer of type `T` and size `size`
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	if size <= 0 {
		size = defaultBufferSize
	}
	return &RingBuffer[T]{
		items: make([]T, size),
	}
}
