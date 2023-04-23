package window

var bartlettMap = map[int]Window{
	8:    Bartlett8,
	16:   Bartlett16,
	32:   Bartlett32,
	64:   Bartlett64,
	128:  Bartlett128,
	256:  Bartlett256,
	512:  Bartlett512,
	1024: Bartlett1024,
	2048: Bartlett2048,
	4096: Bartlett4096,
	8192: Bartlett8192,
}

// Bartlett returns an L-point Bartlett window.
// Reference: http://www.mathworks.com/help/signal/ref/bartlett.html
func Bartlett(i int) Window {
	w, ok := bartlettMap[i]
	if !ok {
		return newBartlett(i)
	}
	return w
}

func newBartlett(i int) Window {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			indices     = float64(i - 1)
			coefficient = 2 / indices
			n           = 0.0
		)

		for ; n <= indices/2; n++ {
			r[int(n)] = coefficient * n
		}
		for ; n <= indices; n++ {
			r[int(n)] = 2 - coefficient*n
		}

		return r
	}
}
