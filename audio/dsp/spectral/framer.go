// Package spectral is the frequency-domain analysis of a single-channel signal of normalised samples: a Framer that
// cuts a stream into overlapping frames, an STFT that turns each frame into a one-sided amplitude spectrum with float
// Hz bins, and fractional-octave bands that reduce a spectrum to a few levels.
//
// Scaling is explicit: a spectrum holds the peak amplitude of a sinusoid centred on each bin, so a full-scale sine
// reads 1 (0 dBFS) on its bin whatever the window, and band levels are corrected for the window's equivalent noise
// bandwidth, so a sine inside a band reads its amplitude as long as the band holds the window's main lobe (a few bins,
// see Banding.Levels).
//
// It is pure DSP on []float64 (CLAUDE.md §5.3): it uses the fft package and the precomputed fft/window tables, and
// nothing beyond the standard library.
package spectral

import (
	"errors"
	"fmt"
)

// ErrInvalidSize is returned for frame sizes, hops, windows and band layouts that cannot be analysed.
var ErrInvalidSize = errors.New("spectral: invalid size")

// Framer cuts a stream of samples into frames of Size samples whose starts lie Hop samples apart, carrying the samples
// of an unfinished frame across calls to Write. Consecutive frames overlap when Hop < Size; when Hop > Size the samples
// between two frames are skipped.
//
// A Framer is not safe for concurrent use.
type Framer struct {
	size, hop int
	buf       []float64 // the start of the next frame: fewer than size samples between calls; grows as samples arrive
	skip      int       // samples still to drop before the next frame starts (Hop > Size)
	pos       int64     // index of buf[0] since the last Reset: the first sample of the next frame
	resets    uint64    // counts Reset calls, so that Write notices one made by its callback
}

// NewFramer returns a Framer of frames of size samples, hop samples apart. Both must be positive. The Framer allocates
// its buffer as samples arrive, up to one frame.
func NewFramer(size, hop int) (*Framer, error) {
	if size < 1 || hop < 1 {
		return nil, fmt.Errorf("%w: frame size %d, hop %d", ErrInvalidSize, size, hop)
	}

	return &Framer{size: size, hop: hop}, nil
}

// Size returns the number of samples in a frame.
func (f *Framer) Size() int { return f.size }

// Hop returns the distance, in samples, between the starts of two consecutive frames.
func (f *Framer) Hop() int { return f.hop }

// Write appends x to the stream and calls fn with every frame it completes, in stream order. The frame is valid only
// during the call and fn must not modify it: overlapping frames share its samples. fn may call Reset, which starts the
// next frame with the rest of x, but not Write. Write never modifies x.
func (f *Framer) Write(x []float64, fn func(frame []float64)) {
	for len(x) > 0 {
		if f.skip > 0 {
			n := min(f.skip, len(x))
			x, f.skip, f.pos = x[n:], f.skip-n, f.pos+int64(n)

			continue
		}

		n := min(f.size-len(f.buf), len(x))

		if need := len(f.buf) + n; need > cap(f.buf) {
			// grow as samples arrive, never beyond one frame: a Framer that is never fed costs nothing
			grown := make([]float64, len(f.buf), min(f.size, max(need, 2*cap(f.buf)))) //nolint:mnd // doubling
			copy(grown, f.buf)
			f.buf = grown
		}

		f.buf, x = append(f.buf, x[:n]...), x[n:]

		if len(f.buf) < f.size {
			return
		}

		resets := f.resets
		fn(f.buf)

		if f.resets != resets {
			continue // fn called Reset: the rest of x starts a new frame
		}

		if f.hop >= f.size {
			f.buf, f.skip, f.pos = f.buf[:0], f.hop-f.size, f.pos+int64(f.size)

			continue
		}

		f.buf, f.pos = f.buf[:copy(f.buf, f.buf[f.hop:])], f.pos+int64(f.hop)
	}
}

// Reset drops the carried samples, so the next Write starts a new frame: call it when the stream is discontinuous
// (a restart, a gap, a new sample rate), so that no frame straddles the discontinuity.
func (f *Framer) Reset() {
	f.buf, f.skip, f.pos = f.buf[:0], 0, 0
	f.resets++
}

// Position returns the index of the first sample of the next frame, counted from the last Reset (or from NewFramer):
// during a Write callback, the first sample of the frame being delivered.
func (f *Framer) Position() int64 { return f.pos }
