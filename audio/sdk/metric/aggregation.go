package metric

import "math"

// Aggregation combines the values an Instrument records between two collections. Record receives each value with the
// number of frames it covers, so averages weigh a stream's short last buffer by its length. Collect returns the combined
// value and resets the aggregation; it reports false when nothing was recorded. Implementations need no locking: the
// Meter serialises every call.
type Aggregation interface {
	Record(v float64, frames int)
	Collect() (float64, bool)
}

// Last keeps the most recent value.
func Last() Aggregation { return &last{} }

// Max keeps the largest value.
func Max() Aggregation { return &extreme{better: func(a, b float64) bool { return a > b }} }

// Min keeps the smallest value.
func Min() Aggregation { return &extreme{better: func(a, b float64) bool { return a < b }} }

// Mean averages the values, weighted by the frames each one covers.
func Mean() Aggregation { return &mean{} }

// Sum adds the values up.
func Sum() Aggregation { return &sum{} }

type last struct {
	v  float64
	ok bool
}

func (a *last) Record(v float64, _ int) { a.v, a.ok = v, true }

func (a *last) Collect() (float64, bool) {
	v, ok := a.v, a.ok
	*a = last{}

	return v, ok
}

type extreme struct {
	better func(a, b float64) bool
	v      float64
	ok     bool
}

func (a *extreme) Record(v float64, _ int) {
	if !a.ok || a.better(v, a.v) || math.IsNaN(v) {
		a.v, a.ok = v, true
	}
}

func (a *extreme) Collect() (float64, bool) {
	v, ok := a.v, a.ok
	a.v, a.ok = 0, false

	return v, ok
}

type mean struct {
	sum    float64
	frames int64
}

func (a *mean) Record(v float64, frames int) {
	w := int64(max(frames, 1))
	a.sum += v * float64(w)
	a.frames += w
}

func (a *mean) Collect() (float64, bool) {
	if a.frames == 0 {
		return 0, false
	}

	v := a.sum / float64(a.frames)
	*a = mean{}

	return v, true
}

type sum struct {
	v  float64
	ok bool
}

func (a *sum) Record(v float64, _ int) { a.v, a.ok = a.v+v, true }

func (a *sum) Collect() (float64, bool) {
	v, ok := a.v, a.ok
	*a = sum{}

	return v, ok
}

// SpectrumAggregation combines, bin by bin, the spectra a spectral Instrument records between two collections. Record
// receives each spectrum, valid only during the call, with the number of frames it covers; every spectrum recorded
// between two collections has the same number of bins (a spectrum of another length starts over). Collect returns the
// combined spectrum in a new slice and resets the aggregation; it reports false when nothing was recorded.
// Implementations need no locking: the Meter serialises every call.
type SpectrumAggregation interface {
	Record(v []float64, frames int)
	Collect() ([]float64, bool)
}

// PeakHold keeps the largest value of every bin.
func PeakHold() SpectrumAggregation { return &peakHold{} }

// MeanSpectrum averages the power of every bin: it reports sqrt(Σ frames·v² / Σ frames), the energy average of the
// amplitudes (as Welch's method does), weighted by the frames each spectrum covers.
func MeanSpectrum() SpectrumAggregation { return &meanSpectrum{} }

type peakHold struct {
	v  []float64
	ok bool
}

func (a *peakHold) Record(v []float64, _ int) {
	if !a.ok || len(v) != len(a.v) {
		a.v, a.ok = append(a.v[:0], v...), true

		return
	}

	for i, x := range v {
		if x > a.v[i] || math.IsNaN(x) {
			a.v[i] = x
		}
	}
}

func (a *peakHold) Collect() ([]float64, bool) {
	if !a.ok {
		return nil, false
	}

	out := append([]float64(nil), a.v...)
	a.v, a.ok = a.v[:0], false

	return out, true
}

// meanSpectrum keeps, per bin, the largest magnitude seen and the weighted sum of squares scaled by it, so that no
// square overflows or underflows.
type meanSpectrum struct {
	peak   []float64
	sum    []float64 // Σ frames·(v/peak)²
	frames int64
}

func (a *meanSpectrum) Record(v []float64, frames int) {
	if a.frames == 0 || len(v) != len(a.peak) {
		a.peak, a.sum, a.frames = make([]float64, len(v)), make([]float64, len(v)), 0
	}

	w := float64(max(frames, 1))
	a.frames += int64(max(frames, 1))

	for i, x := range v {
		x = math.Abs(x)

		switch {
		case math.IsNaN(x) || math.IsNaN(a.sum[i]):
			a.sum[i] = math.NaN()
		case math.IsInf(x, 0) || math.IsInf(a.peak[i], 0):
			a.peak[i] = math.Inf(1)
		case x > a.peak[i]:
			r := a.peak[i] / x
			a.sum[i] = a.sum[i]*r*r + w
			a.peak[i] = x
		case x > 0:
			r := x / a.peak[i]
			a.sum[i] += w * r * r
		}
	}
}

func (a *meanSpectrum) Collect() ([]float64, bool) {
	if a.frames == 0 {
		return nil, false
	}

	out := make([]float64, len(a.peak))

	for i, p := range a.peak {
		switch {
		case math.IsNaN(a.sum[i]):
			out[i] = math.NaN()
		case math.IsInf(p, 0):
			out[i] = p
		default:
			out[i] = p * math.Sqrt(a.sum[i]/float64(a.frames))
		}
	}

	a.frames = 0

	return out, true
}
