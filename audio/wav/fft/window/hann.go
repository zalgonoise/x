package window

import (
	"math"
)

var hannMap = map[int]Window{
	8:    Hann8,
	16:   Hann16,
	32:   Hann32,
	64:   Hann64,
	128:  Hann128,
	256:  Hann256,
	512:  Hann512,
	1024: Hann1024,
	2048: Hann2048,
	4096: Hann4096,
	8192: Hann8192,
}

// Hann returns an L-point Hann window.
// Reference: http://www.mathworks.com/help/signal/ref/hann.html
func Hann(i int) Window {
	w, ok := hannMap[i]
	if !ok {
		return newHann(i)
	}
	return w
}

func newHann(i int) Window {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			indices     = float64(i - 1)
			coefficient = tau / indices
		)

		for n := 0.0; n <= indices; n++ {
			r[int(n)] = 0.5 * (1 - math.Cos(coefficient*n))
		}
		return r
	}
}
