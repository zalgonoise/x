package trig

import "math"

// specialInputs are the values where a range-reduced Sin or Cos most easily diverges from the math package: signed
// zeros, NaN, infinities, the extremes of float64, and both sides of the Payne-Hanek reduction threshold.
func specialInputs() []float64 {
	return []float64{
		0, math.Copysign(0, -1), math.NaN(), math.Inf(1), math.Inf(-1),
		math.MaxFloat64, -math.MaxFloat64, math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64,
		reduceThreshold - 1, reduceThreshold, reduceThreshold + 1, -reduceThreshold,
		PiBy4, math.Pi / 2, math.Pi, 2 * math.Pi, 1e300, -1e-300,
	}
}

// sameFloat reports whether a and b are the same float64, bit for bit; any two NaNs are the same.
func sameFloat(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.IsNaN(a) && math.IsNaN(b)
	}

	return math.Float64bits(a) == math.Float64bits(b)
}
