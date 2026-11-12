package spectral

import (
	"fmt"
	"math"
	"math/cmplx"
	"slices"

	"github.com/zalgonoise/x/audio/fft"
)

// Frame sizes an STFT accepts: powers of two in [MinSize, MaxSize].
const (
	MinSize = 4
	MaxSize = 1 << 16
)

// STFT is a short-time Fourier transform: it frames a stream (see Framer), windows each frame and transforms it into a
// one-sided amplitude spectrum of Size/2+1 bins, bin k centred on k·rate/Size Hz (see BinFrequency).
//
// An STFT is not safe for concurrent use.
type STFT struct {
	framer *Framer
	window []float64 // owned copy, never modified (Clone shares it); nil is rectangular
	scale  float64   // 1/Σw: turns |X[k]| into the amplitude of a sinusoid on bin k (doubled off DC and Nyquist)
	enbw   float64   // equivalent noise bandwidth, in bins

	in  []complex128 // allocated by the first Spectrum call
	out []float64
}

// NewSTFT returns an STFT of size-point frames, hop samples apart, shaped by window: one of the fft/window tables
// (window.Hann(size), for example), which must hold size finite values with a positive sum. A nil window is
// rectangular. The STFT keeps its own copy of the window.
func NewSTFT(size, hop int, window []float64) (*STFT, error) {
	if size < MinSize || size > MaxSize || size&(size-1) != 0 {
		return nil, fmt.Errorf("%w: frame size %d (want a power of two in [%d, %d])", ErrInvalidSize, size, MinSize,
			MaxSize)
	}

	framer, err := NewFramer(size, hop)
	if err != nil {
		return nil, err
	}

	// the sums of window/peak: the ENBW is their ratio, and neither overflows nor underflows whatever the window's scale
	sum, sumSquares, peak := float64(size), float64(size), 1.0

	if window != nil {
		if len(window) != size {
			return nil, fmt.Errorf("%w: window of %d values for %d-point frames", ErrInvalidSize, len(window), size)
		}

		peak, sum, sumSquares = 0, 0, 0

		for _, w := range window {
			peak = max(peak, math.Abs(w))
		}

		for _, w := range window {
			sum += w / peak
			sumSquares += (w / peak) * (w / peak)
		}

		if g := sum * peak; math.IsNaN(sum) || sum <= 0 || math.IsInf(g, 0) || math.IsInf(1/g, 0) {
			return nil, fmt.Errorf("%w: window sum %g (want a positive, finite window)", ErrInvalidSize, sum*peak)
		}

		window = slices.Clone(window)
	}

	return &STFT{
		framer: framer,
		window: window,
		scale:  1 / (sum * peak),
		enbw:   float64(size) * sumSquares / (sum * sum),
	}, nil
}

// Clone returns an independent copy of s, carried samples included. The window, which never changes, is shared, so a
// fresh STFT used only as a template clones cheaply into one STFT per channel.
func (s *STFT) Clone() *STFT {
	c := *s
	f := *s.framer
	f.buf = slices.Clone(f.buf)
	c.framer, c.in, c.out = &f, nil, nil

	return &c
}

// Size returns the number of samples in a frame.
func (s *STFT) Size() int { return s.framer.size }

// Hop returns the distance, in samples, between the starts of two consecutive frames.
func (s *STFT) Hop() int { return s.framer.hop }

// Bins returns the number of bins of each spectrum: Size/2+1, from 0 Hz to the Nyquist frequency.
func (s *STFT) Bins() int { return s.framer.size/2 + 1 }

// ENBW returns the equivalent noise bandwidth of the window, in bins: 1 for a rectangular window, 1.5 for Hann. It
// relates amplitudes to power: the power of a band is Σ A[k]² / (2·ENBW) over its bins.
func (s *STFT) ENBW() float64 { return s.enbw }

// Write appends x to the stream and calls fn with the spectrum of every frame it completes (see Spectrum). The
// spectrum is valid only during the call. Write never modifies x.
func (s *STFT) Write(x []float64, fn func(spectrum []float64)) {
	s.framer.Write(x, func(frame []float64) { fn(s.Spectrum(frame)) })
}

// Reset drops the carried samples (see Framer.Reset).
func (s *STFT) Reset() { s.framer.Reset() }

// Position returns the index of the first sample of the next frame since the last Reset; during a Write callback, that
// of the frame whose spectrum is being delivered (see Framer.Position).
func (s *STFT) Position() int64 { return s.framer.Position() }

// Spectrum returns the one-sided amplitude spectrum of one frame of Size samples: bin k holds the peak amplitude of a
// sinusoid centred on it, so a full-scale sine reads 1. Bins 0 (DC) and Size/2 (Nyquist) hold the mean and the
// alternating component, unscaled. The result is reused by the next call; frame is never modified. A frame of the wrong
// length returns nil.
func (s *STFT) Spectrum(frame []float64) []float64 {
	if len(frame) != s.framer.size {
		return nil
	}

	if s.in == nil {
		s.in, s.out = make([]complex128, s.framer.size), make([]float64, s.Bins())
	}

	for i, v := range frame {
		if s.window != nil {
			v *= s.window[i]
		}

		s.in[i] = complex(v, 0)
	}

	spectrum := fft.FFT(s.in)
	last := len(s.out) - 1

	for k := range s.out {
		a := cmplx.Abs(spectrum[k]) * s.scale
		if k != 0 && k != last {
			a *= 2 // the energy of the mirrored negative-frequency bin
		}

		s.out[k] = a
	}

	return s.out
}

// BinFrequency returns the centre frequency, in Hz, of bin k of a size-point spectrum of audio sampled at rate Hz.
func BinFrequency(k, size, rate int) float64 {
	return float64(k) * float64(rate) / float64(size)
}
