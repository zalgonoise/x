// Package feature computes features of a single-channel signal of normalised samples: levels, crest factor,
// zero-crossing rate, DC offset and clipping. Functions never modify their input and return 0 for an empty signal.
//
// It is pure DSP on []float64 with no dependency beyond the standard library (CLAUDE.md §5.3); the SDK's instruments
// (sdk/metric) call it once per channel.
package feature

import "math"

// Peak returns the largest absolute sample value (linear; 1 is full scale).
func Peak(x []float64) float64 {
	var peak float64

	for _, v := range x {
		peak = max(peak, math.Abs(v))
	}

	return peak
}

// RMS returns the root-mean-square level (linear; 1/√2 for a full-scale sine). Samples are scaled by the peak before
// squaring, so values far outside [-1, 1] (float WAV) neither overflow nor underflow.
func RMS(x []float64) float64 {
	peak := Peak(x)
	if peak == 0 || math.IsInf(peak, 0) || math.IsNaN(peak) {
		return peak
	}

	return peak * scaledRMS(x, peak)
}

// scaledRMS returns the RMS of x / peak, which lies in (0, 1].
func scaledRMS(x []float64, peak float64) float64 {
	var sum float64

	for _, v := range x {
		s := v / peak
		sum += s * s
	}

	return math.Sqrt(sum / float64(len(x)))
}

// Crest returns the peak-to-RMS ratio: 1 for a square wave, √2 for a sine. Silence returns 0.
func Crest(x []float64) float64 {
	peak := Peak(x)

	switch {
	case peak == 0:
		return 0
	case math.IsInf(peak, 0) || math.IsNaN(peak):
		return math.NaN()
	}

	return 1 / scaledRMS(x, peak)
}

// ZeroCrossings counts sign changes between consecutive samples; zero counts as positive.
func ZeroCrossings(x []float64) int {
	n := 0

	for i := 1; i < len(x); i++ {
		if (x[i-1] < 0) != (x[i] < 0) {
			n++
		}
	}

	return n
}

// ZeroCrossingRate returns the zero crossings per second of a signal sampled at rate Hz. Signals shorter than two
// samples, or a rate of zero or less, return 0.
func ZeroCrossingRate(x []float64, rate int) float64 {
	if len(x) < 2 || rate <= 0 {
		return 0
	}

	return float64(ZeroCrossings(x)) * float64(rate) / float64(len(x)-1)
}

// DCOffset returns the mean sample value.
func DCOffset(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}

	var sum float64

	for _, v := range x {
		sum += v
	}

	return sum / float64(len(x))
}

// Clipped counts the samples whose magnitude is at least threshold (1 for full scale; decoded integer PCM reaches
// -1 but at most 1-2^-(n-1), so a threshold slightly below 1 catches positive clipping too).
func Clipped(x []float64, threshold float64) int {
	n := 0

	for _, v := range x {
		if math.Abs(v) >= threshold {
			n++
		}
	}

	return n
}

// DBFS converts a linear level to decibels relative to full scale (20·log10). Zero returns -Inf.
func DBFS(level float64) float64 {
	const amplitudeDB = 20

	return amplitudeDB * math.Log10(level)
}
