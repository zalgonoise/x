package file

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
		if err = w.open(); err != nil {
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

func (e *writer) open() error {
	if e.dir == "" {
		e.dir = os.TempDir()
	}

	if e.name == "" {
		e.name = defaultSlug
	}

	target := fmt.Sprintf("%s/%s_%04d", e.dir, e.name, e.idx)

	f, err := os.Open(target)

	if errors.Is(err, fs.ErrNotExist) {
		f, err = os.Create(target)
	}

	if err != nil {
		return err
	}

	e.idx++
	e.w = f

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
