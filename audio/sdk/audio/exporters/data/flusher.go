package data

import (
	"time"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

type Flusher struct {
	data []byte
	len  int
	cap  int

	fn  FlushFunc
	dur time.Duration
}

type FlushFunc func(id string, h *wav.Header, data []byte) error

func NewFlusher(dur time.Duration, fn FlushFunc) *Flusher {
	return &Flusher{
		fn:  fn,
		dur: dur,
	}
}

func (f *Flusher) init(h *wav.Header) {
	bufferSize := wav.TimeToBufferSize(
		wav.ByteRate(h.SampleRate, h.BitsPerSample, h.NumChannels),
		f.dur,
	)

	f.data = make([]byte, bufferSize)
	f.cap = bufferSize
}

func (f *Flusher) Write(id string, h *wav.Header, b []byte) (n int, err error) {
	if f.data == nil {
		f.init(h)
	}

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

	n := f.cap - f.len
	copy(f.data[f.len:], b[:n])

	f.len = f.cap

	b = b[n:]

	switch l := len(b); {
	case l == 0:
		return f.FlushIfFull(id, h)
	case l > f.cap:
		return f.writeOutOfBounds(id, h, b)
	default:
		return f.writeWithinBounds(id, h, b)
	}
}

func (f *Flusher) writeWithinBounds(id string, h *wav.Header, b []byte) error {
	if err := f.FlushIfFull(id, h); err != nil {
		return err
	}

	copy(f.data[f.len:], b)
	f.len += len(b)

	return f.FlushIfFull(id, h)
}

func (f *Flusher) FlushIfFull(id string, h *wav.Header) error {
	if f.data == nil {
		f.init(h)
	}

	if f.len == f.cap {
		if err := f.fn(id, h, f.data[:f.cap]); err != nil {
			return err
		}

		f.len = 0
	}

	return nil
}
