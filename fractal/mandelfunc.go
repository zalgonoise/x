package fractal

import "math/cmplx"

func Mandelfunc(c, z complex128, fn MandelFn) (n int) {
	for ; cmplx.Abs(z) < 2 && n < maxIter; n++ {
		z = fn(c, z)
	}
	return n
}

type MandelFn func(c, z complex128) complex128
