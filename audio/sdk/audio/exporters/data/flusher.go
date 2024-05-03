package data

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
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
	if err := f.ForceFlush(id, h); err != nil {
		return err
	}

	n := f.cap - f.len
	copy(f.data[f.len:], b[:n])

	f.len = f.cap

	b = b[n:]

	switch l := len(b); {
	case l == 0:
		return f.ForceFlush(id, h)
	case l > f.cap:
		return f.writeOutOfBounds(id, h, b)
	default:
		return f.writeWithinBounds(id, h, b)
	}
}

func (f *Flusher) writeWithinBounds(id string, h *wav.Header, b []byte) error {
	if err := f.ForceFlush(id, h); err != nil {
		return err
	}

	copy(f.data[f.len:], b)
	f.len += len(b)

	return f.ForceFlush(id, h)
}

func (f *Flusher) ForceFlush(id string, h *wav.Header) error {
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

type ZLibFlusher struct {
	b   *bytes.Buffer
	w   *zlib.Writer
	len int
	cap int

	fn  FlushFunc
	dur time.Duration
}

func NewZLibFlusher(dur time.Duration, fn FlushFunc) *ZLibFlusher {
	return &ZLibFlusher{
		fn:  fn,
		dur: dur,
	}
}

func (f *ZLibFlusher) init(h *wav.Header) error {
	f.cap = wav.TimeToBufferSize(
		wav.ByteRate(h.SampleRate, h.BitsPerSample, h.NumChannels),
		f.dur,
	)

	f.b = bytes.NewBuffer(make([]byte, 0, f.cap))

	var err error
	if f.w, err = zlib.NewWriterLevel(f.b, zlib.BestCompression); err != nil {
		return err
	}

	return nil
}

func (f *ZLibFlusher) Write(id string, h *wav.Header, b []byte) (n int, err error) {
	if err = f.flush(id, h); err != nil {
		return 0, err
	}

	if n, err = f.w.Write(b); err != nil {
		return n, err
	}

	f.len += n

	return n, f.flush(id, h)
}

func (f *ZLibFlusher) flush(id string, h *wav.Header) error {
	if f.w == nil || f.b == nil {
		if err := f.init(h); err != nil {
			return err
		}
	}

	for f.len > f.cap {
		if err := f.w.Flush(); err != nil {
			return err
		}

		maximum := f.cap
		if f.b.Len() < maximum {
			maximum = f.b.Len()
		}

		data := make([]byte, maximum)

		n, err := f.b.Read(data)
		if err != nil {
			return err
		}

		f.len -= n

		if err := f.fn(id, h, data); err != nil {
			return err
		}
	}

	return nil
}

func (f *ZLibFlusher) ForceFlush(id string, h *wav.Header) error {
	return f.flush(id, h)
}

type GZipFlusher struct {
	b   *bytes.Buffer
	w   *gzip.Writer
	len int
	cap int

	fn  FlushFunc
	dur time.Duration
}

func NewGZipFlusher(dur time.Duration, fn FlushFunc) *GZipFlusher {
	return &GZipFlusher{
		fn:  fn,
		dur: dur,
	}
}

func (f *GZipFlusher) init(h *wav.Header) error {
	f.cap = wav.TimeToBufferSize(
		wav.ByteRate(h.SampleRate, h.BitsPerSample, h.NumChannels),
		f.dur,
	)

	f.b = bytes.NewBuffer(make([]byte, 0, f.cap))

	var err error
	if f.w, err = gzip.NewWriterLevel(f.b, gzip.BestCompression); err != nil {
		return err
	}

	return nil
}

func (f *GZipFlusher) Write(id string, h *wav.Header, b []byte) (n int, err error) {
	if err = f.flush(id, h); err != nil {
		return 0, err
	}

	if n, err = f.w.Write(b); err != nil {
		return n, err
	}

	f.len += n

	return n, f.flush(id, h)
}

func (f *GZipFlusher) flush(id string, h *wav.Header) error {
	if f.w == nil || f.b == nil {
		if err := f.init(h); err != nil {
			return err
		}
	}

	for f.len > f.cap {
		if err := f.w.Flush(); err != nil {
			return err
		}

		maximum := f.cap
		if f.b.Len() < maximum {
			maximum = f.b.Len()
		}

		data := make([]byte, maximum)

		n, err := f.b.Read(data)
		if err != nil {
			return err
		}

		f.len -= n

		if err := f.fn(id, h, data); err != nil {
			return err
		}
	}

	return nil
}

func (f *GZipFlusher) ForceFlush(id string, h *wav.Header) error {
	return f.flush(id, h)
}
