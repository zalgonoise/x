package audio

import (
	"slices"
	"time"
)

// Buffer is a block of decoded audio: interleaved samples normalised to [-1, 1).
//
// A Buffer received by a component (an exporter, an extractor) is read-only and valid only for the duration of the
// call that received it: anything kept past the call must be copied with Clone.
type Buffer struct {
	Format  Format
	Samples []float64 // interleaved; len(Samples) is a multiple of Format.Channels
	Frame   int64     // index of the first frame since the start of the stream
}

// Frames returns the number of whole frames in the buffer.
func (b Buffer) Frames() int {
	if b.Format.Channels <= 0 {
		return 0
	}

	return len(b.Samples) / b.Format.Channels
}

// Duration returns how long the buffer lasts.
func (b Buffer) Duration() time.Duration { return b.Format.Duration(int64(b.Frames())) }

// Offset returns the time of the buffer's first frame since the start of the stream.
func (b Buffer) Offset() time.Duration { return b.Format.Duration(b.Frame) }

// Channel appends channel i (0-based) of every frame to dst and returns it. An out-of-range channel returns dst.
func (b Buffer) Channel(i int, dst []float64) []float64 {
	channels := b.Format.Channels
	if i < 0 || i >= channels {
		return dst
	}

	dst = slices.Grow(dst, b.Frames())

	for f := range b.Frames() {
		dst = append(dst, b.Samples[f*channels+i])
	}

	return dst
}

// Mixdown appends the average of all channels of every frame to dst and returns it. Mono buffers are copied as-is; a
// buffer without channels has no frames and appends nothing.
func (b Buffer) Mixdown(dst []float64) []float64 {
	channels := b.Format.Channels

	switch {
	case channels <= 0:
		return dst
	case channels == 1:
		return append(dst, b.Samples...)
	}

	dst = slices.Grow(dst, b.Frames())

	for f := range b.Frames() {
		var sum float64

		for _, v := range b.Samples[f*channels : (f+1)*channels] {
			sum += v
		}

		dst = append(dst, sum/float64(channels))
	}

	return dst
}

// Clone returns a copy of the buffer that does not share its samples.
func (b Buffer) Clone() Buffer {
	b.Samples = slices.Clone(b.Samples)

	return b
}
