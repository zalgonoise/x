package testutil

import (
	"math"
	"math/rand/v2"
)

// Sine returns n samples of amp * sin(2π * hz * t + phase), sampled at rate Hz.
func Sine(n, rate int, hz, amp, phase float64) []float64 {
	out := make([]float64, n)

	for i := range out {
		out[i] = amp * math.Sin(2*math.Pi*hz*float64(i)/float64(rate)+phase)
	}

	return out
}

// Cosine returns n samples of amp * cos(2π * hz * t + phase), sampled at rate Hz.
func Cosine(n, rate int, hz, amp, phase float64) []float64 {
	return Sine(n, rate, hz, amp, phase+math.Pi/2)
}

// BinCosine returns n samples of amp * cos(2π * k * i / n): a cosine centred exactly on FFT bin k of an n-point
// transform, so its spectrum has closed-form values (amp*n/2 at bins k and n-k with a rectangular window).
func BinCosine(n, k int, amp float64) []float64 {
	out := make([]float64, n)

	for i := range out {
		out[i] = amp * math.Cos(2*math.Pi*float64(k)*float64(i)/float64(n))
	}

	return out
}

// Impulse returns n samples that are zero except for a 1.0 at index at.
func Impulse(n, at int) []float64 {
	out := make([]float64, n)

	if at >= 0 && at < n {
		out[at] = 1
	}

	return out
}

// DC returns n samples set to v.
func DC(n int, v float64) []float64 {
	out := make([]float64, n)

	for i := range out {
		out[i] = v
	}

	return out
}

// Alternating returns n samples of (-1)^i, the Nyquist-frequency signal.
func Alternating(n int) []float64 {
	out := make([]float64, n)

	for i := range out {
		if i%2 == 0 {
			out[i] = 1
		} else {
			out[i] = -1
		}
	}

	return out
}

// Noise returns n uniformly distributed samples in [-amp, amp), deterministic for a given seed.
func Noise(n int, amp float64, seed uint64) []float64 {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	out := make([]float64, n)

	for i := range out {
		out[i] = amp * (2*rng.Float64() - 1)
	}

	return out
}

// Chirp returns n samples of a linear sine sweep from f0 to f1 Hz, sampled at rate Hz, with amplitude amp.
func Chirp(n, rate int, f0, f1, amp float64) []float64 {
	out := make([]float64, n)
	duration := float64(n) / float64(rate)
	k := (f1 - f0) / duration

	for i := range out {
		t := float64(i) / float64(rate)
		out[i] = amp * math.Sin(2*math.Pi*(f0*t+k*t*t/2))
	}

	return out
}

// Interleave merges per-channel sample slices into one interleaved slice (L0 R0 L1 R1 ...). The output length is
// the length of the shortest channel times the number of channels.
func Interleave[T any](channels ...[]T) []T {
	if len(channels) == 0 {
		return nil
	}

	frames := len(channels[0])
	for _, ch := range channels[1:] {
		frames = min(frames, len(ch))
	}

	out := make([]T, 0, frames*len(channels))

	for i := range frames {
		for _, ch := range channels {
			out = append(out, ch[i])
		}
	}

	return out
}

// Deinterleave splits an interleaved slice into numChannels per-channel slices, dropping a trailing partial frame.
func Deinterleave[T any](samples []T, numChannels int) [][]T {
	if numChannels <= 0 {
		return nil
	}

	frames := len(samples) / numChannels
	out := make([][]T, numChannels)

	for c := range out {
		out[c] = make([]T, frames)

		for i := range frames {
			out[c][i] = samples[i*numChannels+c]
		}
	}

	return out
}

// IntExtremes returns the boundary values of a signed bits-wide integer: min, min+1, -1, 0, 1, max-1, max.
func IntExtremes(bits int) []int32 {
	maxV := int32(1<<(bits-1) - 1)
	minV := -maxV - 1

	return []int32{minV, minV + 1, -1, 0, 1, maxV - 1, maxV}
}
