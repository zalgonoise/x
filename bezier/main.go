package main

import (
	"bytes"
	"image"
	"image/png"
	"os"

	svg "github.com/ajstarks/svgo"
	"gocv.io/x/gocv"
)

func main() {
	// Open the PNG file
	pngFile, err := os.Open("image.png")
	if err != nil {
		panic(err)
	}
	defer pngFile.Close()

	// Decode the PNG image
	pngImage, err := png.Decode(pngFile)
	if err != nil {
		panic(err)
	}

	// Convert the image to grayscale
	bounds := pngImage.Bounds()
	grayImage := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImage.Set(x, y, pngImage.At(x, y))
		}
	}

	// Encode the grayscale image into a byte slice
	grayData := &bytes.Buffer{}
	if err := png.Encode(grayData, grayImage); err != nil {
		panic(err)
	}

	// Create a gocv.Mat from the byte slice
	grayMat, err := gocv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), gocv.MatTypeCV8UC1, grayData.Bytes())
	if err != nil {
		panic(err)
	}

	// Detect edges in the grayscale image using the Canny edge detector
	cannyImage := gocv.NewMat()
	gocv.Canny(grayMat, &cannyImage, 50, 150)

	// Find contours in the canny image
	contours := gocv.FindContours(cannyImage, gocv.RetrievalExternal, gocv.ChainApproxNone)

	// Open the SVG file
	svgFile, err := os.Create("image.svg")
	if err != nil {
		panic(err)
	}
	defer svgFile.Close()

	// Create a new SVG canvas
	canvas := svg.New(svgFile)

	// Start the SVG canvas with the dimensions of the input image
	canvas.Start(bounds.Dx(), bounds.Dy())

	// Set the stroke width and color for the SVG elements
	// strokeWidth := 2
	// strokeColor := "black"

	// Draw the detected contours on the SVG canvas
	for _, contour := range contours.ToPoints() {
		for i := 0; i < len(contour); i++ {
			x1, y1 := contour[i].X, contour[i].Y
			x2, y2 := contour[(i+1)%len(contour)].X, contour[(i+1)%len(contour)].Y
			canvas.Line(x1, y1, x2, y2) //fmt.Sprintf("stroke-width:%d;stroke:%s", strokeWidth, strokeColor))
		}
	}

	// End the SVG canvas
	canvas.End()
}
