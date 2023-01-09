package fractal

import (
	"image/png"
	"os"
	"testing"
)

func TestColorMap(t *testing.T) {
	// Open a new PNG file
	f, err := os.Create("colors.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img := DrawMandelbrot(1080, 720)

	// Encode the image to PNG and write it to the file
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
