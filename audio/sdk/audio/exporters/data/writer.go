package data

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

const defaultSlug = "audio-capture"

type writer struct {
	w io.WriteCloser

	dir  string
	name string
	idx  int
}

func (w *writer) Write(p []byte) (n int, err error) {
	if w.w == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	return w.w.Write(p)
}

func (w *writer) Close() error {
	if w.w == nil {
		return nil
	}

	err := w.w.Close()

	w.w = nil

	return err
}

func (w *writer) Reset() error {
	if w.w != nil {
		if err := w.Close(); err != nil {
			return err
		}
	}

	return w.open()
}

func (w *writer) open() error {
	if w.dir == "" {
		w.dir = os.TempDir()
	}

	if w.name == "" {
		w.name = defaultSlug
	}

	target := fmt.Sprintf("%s/%s_%04d", w.dir, w.name, w.idx)

	f, err := os.Open(target)

	if errors.Is(err, fs.ErrNotExist) {
		f, err = os.Create(target)
	}

	if err != nil {
		return err
	}

	w.idx++
	w.w = f

	return nil
}

func NewWriter(path, prefix string) (io.WriteCloser, error) {
	w := &writer{
		dir:  path,
		name: prefix,
		idx:  1,
	}

	if err := w.open(); err != nil {
		return nil, err
	}

	return w, nil
}
