package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"strconv"
	"strings"

	"image/jpeg"
	"image/png"
)

type conf struct {
	path   string
	cols   int
	rows   int
	format string
}

func flags() conf {
	f := flag.String("f", "", "file to crop")
	c := flag.Int("c", 2, "number of columns to crop")
	r := flag.Int("r", 2, "number of rows to crop")
	flag.Parse()

	return conf{
		path: *f,
		cols: *c,
		rows: *r,
	}
}

func main() {
	cfg := flags()
	// open the original image
	i, err := os.Open(cfg.path)
	if err != nil {
		panic(err)
	}

	// decode the image to set format
	img, format, e := image.Decode(i)
	if e != nil {
		fmt.Printf("image decode: %v\n", e)
		os.Exit(1)
	}
	cfg.format = format

	// crop the image
	err = crop(img, cfg)
	if err != nil {
		fmt.Printf("image crop: %v\n", err)
		os.Exit(1)
	}
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

func crop(img image.Image, cfg conf) error {
	bounds := img.Bounds()

	// set the width and height of each tile
	tileWidth := bounds.Max.X / cfg.cols
	tileHeight := bounds.Max.Y / cfg.rows
	nameIdx := cfg.cols * cfg.rows
	baseName := extractExt(cfg.path, cfg.format)

	for row := 0; row < cfg.rows; row++ {
		for col := 0; col < cfg.cols; col++ {
			// set filename
			filename := new(strings.Builder)
			filename.WriteString(baseName)
			filename.WriteByte('_')
			filename.WriteString(strconv.Itoa(nameIdx))
			filename.WriteByte('.')
			filename.WriteString(cfg.format)
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

			switch cfg.format {
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
