// This will be an exported Android Archive (.aar) to be used in Flutter
//
// To build the .aar file, run: `go build -buildmode=c-shared -o process.aar process.go`
// To build the .h / .so files, run: `go build -buildmode=c-shared -o process.so process.go`
package main

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
)

//export ProcessImage
func ProcessImage(path string, cols, rows int, data []byte) error {
	if path == "" || len(data) == 0 {
		return errors.New("invalid image path or content")
	}
	if cols == 0 || rows == 0 || cols*rows == 1 {
		return errors.New("invalid split values")
	}

	buf := bytes.NewBuffer(data)
	img, format, err := image.Decode(buf)
	if err != nil {
		return err
	}

	return crop(img, format, path, cols, rows)
}

func crop(img image.Image, format, path string, cols, rows int) error {
	bounds := img.Bounds()

	// set the width and height of each tile
	tileWidth := bounds.Max.X / cols
	tileHeight := bounds.Max.Y / rows
	nameIdx := cols * rows
	baseName := extractExt(path, format)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			// set filename
			filename := new(strings.Builder)
			filename.WriteString(baseName)
			filename.WriteByte('_')
			filename.WriteString(strconv.Itoa(nameIdx))
			filename.WriteByte('.')
			filename.WriteString(format)
			nameIdx--

			// calculate the bounds of the current tile
			tileBounds := image.Rect(
				col*tileWidth,
				row*tileHeight,
				(col+1)*tileWidth,
				(row+1)*tileHeight,
			)

			// create a subimage from the original image
			tile := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(tileBounds)

			// save the new image
			tileFile, err := os.Create(filename.String())
			if err != nil {
				return err
			}
			defer tileFile.Close()

			switch format {
			case "jpeg":
				if err := encodeJPEG(tileFile, tile); err != nil {
					return err
				}
			case "png":
				if err := encodePNG(tileFile, tile); err != nil {
					return err
				}
			default:
				return errors.New("invalid format")
			}
		}
	}
	return nil
}

func extractExt(path, format string) string {
	return path[:len(path)-len(format)]
}
func encodePNG(w io.Writer, m image.Image) error {
	return png.Encode(w, m)
}
func encodeJPEG(w io.Writer, m image.Image) error {
	return jpeg.Encode(w, m, nil)
}

func main() {}
