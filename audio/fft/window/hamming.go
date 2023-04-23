package window

import "github.com/zalgonoise/x/audio/trig"

var hammingMap = map[int]Window{
	8:    Hamming8,
	16:   Hamming16,
	32:   Hamming32,
	64:   Hamming64,
	128:  Hamming128,
	256:  Hamming256,
	512:  Hamming512,
	1024: Hamming1024,
	2048: Hamming2048,
	4096: Hamming4096,
	8192: Hamming8192,
}

// Hamming returns an L-point symmetric Hamming window.
// Reference: http://www.mathworks.com/help/signal/ref/hamming.html
func Hamming(i int) Window {
	w, ok := hammingMap[i]
	if !ok {
		return newHamming(i)
	}
	return w
}

func newHamming(i int) Window {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			idx         = float64(i - 1)
			coefficient = tau / idx
		)

		for n := 0.0; n <= idx; n++ {
			r[int(n)] = 0.54 - 0.46*trig.Cos(coefficient*n)
		}

		return r
	}
}
