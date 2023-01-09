package fractal

import "math/cmplx"

const maxIter = 10000000

func Mandelbrot(z complex128) (n int) {
	var c complex128 = z
	for ; cmplx.Abs(z) < 2 && n < maxIter; n++ {
		z = mandelbrot(c, z)
	}
	return n
}

func mandelbrot(c, z complex128) complex128 {
	return z*z + c
}
