package data

import "github.com/zalgonoise/x/audio/encoding/wav"

type Flusher struct {
	data []byte
	len  int
	cap  int

	fn FlushFunc
}

type FlushFunc func(id string, h *wav.Header, data []byte) error

func NewFlusher(size int, fn FlushFunc) *Flusher {
	return &Flusher{
		data: make([]byte, size),
		cap:  size,
		fn:   fn,
	}
}

func (f *Flusher) Write(id string, h *wav.Header, b []byte) (n int, err error) {
	switch {
	case len(b) <= f.cap-f.len:
		if err := f.writeWithinBounds(id, h, b); err != nil {
			return 0, err
		}

		return len(b), nil

	case f.cap-f.len == 0:
		if err := f.fn(id, h, f.data[:]); err != nil {
			return 0, err
		}

		return f.Write(id, h, b)
	default:
		if err := f.writeOutOfBounds(id, h, b); err != nil {
			return 0, err
		}

		return len(b), nil
	}
}

func (f *Flusher) writeOutOfBounds(id string, h *wav.Header, b []byte) error {
	if err := f.FlushIfFull(id, h); err != nil {
		return err
	}

	var idx int

	for ; idx < f.cap-f.len; idx++ {
		f.data[f.len+idx] = b[idx]
	}

	if idx < len(b) {
		return f.writeOutOfBounds(id, h, b[idx:])
	}

	return f.FlushIfFull(id, h)
}

func (f *Flusher) writeWithinBounds(id string, h *wav.Header, b []byte) error {
	if err := f.FlushIfFull(id, h); err != nil {
		return err
	}

	for i := range b {
		f.data[f.len+i] = b[i]
	}

	f.len += len(b)

	return f.FlushIfFull(id, h)
}

func (f *Flusher) FlushIfFull(id string, h *wav.Header) error {
	if f.len == f.cap {
		if err := f.fn(id, h, f.data[:f.cap]); err != nil {
			return err
		}

		f.len = 0
	}

	return nil
}
