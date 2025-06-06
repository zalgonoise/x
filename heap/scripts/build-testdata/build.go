package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"math/rand"
	"os"
	"path"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	n := flag.Int("n", 500, "number of integers to generate")
	o := flag.String("o", "./testdata/integers.json", "output file")
	flag.Parse()

	slice := make([]int, *n)
	r := rand.New(rand.NewSource(0))
	for i := range slice {
		slice[i] = r.Int()
	}

	buf, err := json.Marshal(slice)
	if err != nil {
		logger.ErrorContext(context.Background(), "encoding data as JSON", slog.String("error", err.Error()))

		os.Exit(1)
	}

	dir, _ := path.Split(*o)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
			logger.ErrorContext(context.Background(), "creating output directory", slog.String("error", err.Error()))

			os.Exit(1)
		}
	}

	f, err := os.Create(*o)
	if err != nil {
		logger.ErrorContext(context.Background(), "creating file", slog.String("error", err.Error()))

		os.Exit(1)
	}

	if _, err := f.Write(buf); err != nil {
		logger.ErrorContext(context.Background(), "writing file", slog.String("error", err.Error()))

		os.Exit(1)
	}

	if err := f.Close(); err != nil {
		logger.ErrorContext(context.Background(), "closing file", slog.String("error", err.Error()))

		os.Exit(1)
	}
}
