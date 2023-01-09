package fractal

import (
	"image"
	"image/draw"
)

const (
	xmin, ymin, xmax, ymax = -2, -2, 2, 2
)

func DrawMandelbrot(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Iterate over the pixels in the image
	for py := 0; py < height; py++ {
		y := float64(py)/float64(height)*(ymax-ymin) + ymin
		for px := 0; px < width; px++ {
			x := float64(px)/float64(width)*(xmax-xmin) + xmin
			z := complex(x, y)
			// Calculate the Mandelbrot value for the pixel
			n := Mandelbrot(z)
			// Set the color of the pixel based on the Mandelbrot value
			img.Set(px, py, colorMap(n))
		}
	}

	return img
}

func DrawJulia(width, height int, c complex128) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Iterate over the pixels in the image
	for py := 0; py < height; py++ {
		y := float64(py)/float64(height)*(ymax-ymin) + ymin
		for px := 0; px < width; px++ {
			x := float64(px)/float64(width)*(xmax-xmin) + xmin
			z := complex(x, y)
			// Calculate the Mandelbrot value for the pixel
			n := Julia(z, c)
			// Set the color of the pixel based on the Mandelbrot value
			img.Set(px, py, colorMap(n))
		}
	}

	return img
}

func DrawFn(width, height int, c complex128, fn MandelFn) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Iterate over the pixels in the image
	for py := 0; py < height; py++ {
		y := float64(py)/float64(height)*(ymax-ymin) + ymin
		for px := 0; px < width; px++ {
			x := float64(px)/float64(width)*(xmax-xmin) + xmin
			z := complex(x, y)
			// Calculate the Mandelbrot value for the pixel
			n := Mandelfunc(c, z, fn)
			// Set the color of the pixel based on the Mandelbrot value
			img.Set(px, py, colorMap(n))
		}
	}

	return img
}
