package gbuf

import (
	"errors"
	"io"

	"github.com/zalgonoise/gio"
)

// RingFilter is a buffer that is connected end-to-end, which allows continuous
// reads and writes where the caller configures a callback function to process
// all items on each loop
type RingFilter[T any] struct {
	write int
	read  int
	fn    func([]T) error
	items []T
}

// Write sets the contents of `p` to the buffer, in sequential order.
// The return value n is the length of p; err comes from the process func.
// If the index in the buffer has not been yet read, the entire unread
// buffer value is sent to the configured process function
func (r *RingFilter[T]) Write(p []T) (n int, err error) {
	var (
		ln     = len(p)
		ringLn = r.Len()
	)

	if ln < ringLn {
		end := r.write + ln
		if end <= len(r.items) {
			n = copy(r.items[r.write:r.write+ln], p)
			err = r.fn(r.items[r.write : r.write+ln])
			r.write = r.write + ln

			if err != nil {
				return n, err
			}

			return n, nil
		}

		end %= len(r.items)

		n = copy(r.items[r.write:len(r.items)], p)
		err = r.fn(r.items[r.write:len(r.items)])

		n += copy(r.items[:end], p[n:])
		err = r.fn(r.items[:end])
		r.write = end

		if err != nil {
			return n, err
		}

		return n, nil
	}

	for n < ln {
		if r.write > 0 {
			ringLn = r.Len()
			if n+ringLn > ln {
				ringLn -= (n + ringLn) % ln
			}

			n += copy(r.items[r.write:len(r.items)], p[n:n+ringLn])
			err = r.fn(r.items[r.write:len(r.items)])
			// buffer was filled successfully
			r.write = 0

			if err != nil {
				return n, err
			}

			continue
		}

		ringLn = r.Cap()
		if n+ringLn > ln {
			ringLn -= (n + ringLn) % ln
		}

		n += copy(r.items[0:ringLn], p[n:n+ringLn])
		err = r.fn(r.items[0:ringLn])
		r.write = (r.write + ringLn) % len(r.items)
		r.read = r.write

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of T items written; it always fits into an
// int, but it is int64 to match the gio.WriterTo interface. Any error
// encountered during the write is also returned.
func (r *RingFilter[T]) WriteTo(b gio.Writer[T]) (n int64, err error) {
	for {
		if r.write < r.read {
			num, err := b.Write(r.items[r.write:r.read])
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
			num, err := b.Write(r.items[r.write:len(r.items)])
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
			num, err = b.Write(r.items[:r.read])
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

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (r *RingFilter[T]) Reset() {
	r.write = 0
	r.read = 0
}

// Read reads the next len(p) T items from the buffer or until the buffer
// is drained. The return value n is the number of T items read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (r *RingFilter[T]) Read(p []T) (n int, err error) {
	var (
		ln     = len(p)
		ringLn = r.Len()
	)

	switch {
	case r.read >= r.write:
		size := len(r.items) - r.read
		n += copy(p[:size], r.items[r.read:len(r.items)])
		r.read = 0

		// don't keep writing if there isn't enough space in p
		if n >= ln {
			return n, nil
		}

		n += copy(p[n:n+r.write], r.items[:r.write+1])
		r.read = r.write % len(r.items)

		return n, nil
	default:
		m := copy(p[:ringLn], r.items[r.read:r.write])
		if ln < m {
			n = ln
		} else {
			n = m
		}
		r.read += n

		return n, nil
	}
}

// ReadFrom reads data from b until EOF and appends it to the buffer, cycling
// the buffer as needed. For each complete cycle, the process function is called
// with the full buffer in the ring, and the ring reset.
// The return value n is the number of T items read. Any error except io.EOF
// encountered during the read is also returned.
func (r *RingFilter[T]) ReadFrom(b gio.Reader[T]) (n int64, err error) {

	var num int
	for {
		num, err = b.Read(r.items[r.read:len(r.items)])
		if n < 0 {
			panic(errNegativeRead)
		}
		switch {
		case errors.Is(err, io.EOF):
			r.read = (r.read + num) % len(r.items)
			err = r.fn(r.items[r.write:len(r.items)])
			if err != nil {
				return n, err
			}
			return n, nil
		case err != nil:
			return n, err
		}

		r.read = (r.read + num) % len(r.items)
		n += int64(num)
		if r.read == r.write {
			err = r.fn(r.items[r.write:len(r.items)])
			if err != nil {
				return n, err
			}
			r.Reset()
		}
		if err != nil {
			return n, err
		}
	}
}

// ReadFromIf extends ReadFrom with a conditional function that is called on each loop.
// If the condition(s) is / are met, the loop will continue.
// The return value n is the number of T items read. Any error except io.EOF
// encountered during the read is also returned.
func (r *RingFilter[T]) ReadFromIf(b gio.Reader[T], cond func() bool) (n int64, err error) {
	for cond() {
		num, err := b.Read(r.items[r.read:len(r.items)])
		if n < 0 {
			panic(errNegativeRead)
		}
		if errors.Is(err, io.EOF) {
			r.read = (r.read + num) % len(r.items)
			err = r.fn(r.items[r.write:len(r.items)])
			if err != nil {
				return n, err
			}
			return n, nil
		}
		r.read = (r.read + num) % len(r.items)
		n += int64(num)
		if r.read == r.write {
			err = r.fn(r.items[r.write:len(r.items)])
			if err != nil {
				return n, err
			}
			r.Reset()
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return n, err
		}
	}
	return n, err
}

// Value returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
func (r *RingFilter[T]) Value() []T {
	var (
		n     int
		items []T
	)
	if r.write == r.read {
		return nil
	}
	if r.write < r.read {
		if r.read < len(r.items) {
			r.read++
		}
		items = make([]T, r.read-r.write)
		n = copy(items, r.items[r.write:r.read])
	} else {
		items = make([]T, len(r.items))
		n = copy(items, r.items[r.write-1:])
		if n < len(items) {
			n += copy(items[n:], r.items[:r.read])
		}
	}
	r.Reset()
	return items
}

// Len returns the number of T items of the unread portion of the buffer;
// b.Len() == len(b.T items()).
func (r *RingFilter[T]) Len() int {
	if r.read < r.write {
		return r.write - r.read
	}
	return len(r.items) - (r.read - r.write)
}

// Cap returns the length of the buffer's underlying T item slice, that is, the
// total ring buffer's capacity.
func (r *RingFilter[T]) Cap() int {
	return len(r.items)
}

// NewRingFilter creates a RingFilter of type `T` and size `size`, with process function `fn`
func NewRingFilter[T any](size int, fn func([]T) error) *RingFilter[T] {
	if size <= 0 {
		size = defaultBufferSize
	}
	return &RingFilter[T]{
		items: make([]T, size),
		fn:    fn,
	}
}
