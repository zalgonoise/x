package recorder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"uuid"
)

// Errors returned by Dir.
var (
	// ErrInvalidName is returned by Put for a name that is not a local path inside the directory.
	ErrInvalidName = errors.New("recorder: invalid object name")
	// ErrShortObject is returned by Put when r holds a different number of bytes than size.
	ErrShortObject = errors.New("recorder: object size mismatch")
)

// File modes for the directories and segment files a Dir creates.
const (
	dirMode  os.FileMode = 0o750
	fileMode os.FileMode = 0o640
)

// Dir is a Store keeping each object as a file in a directory. Writes are atomic: an object is written to a temporary
// file and renamed into place, so readers never see a partial segment. Names are confined to the directory (os.Root):
// "a/b.wav" creates a subdirectory, "../x" is rejected.
type Dir struct {
	root *os.Root
}

// NewDir returns a Store for the directory at path, creating it if needed. Close releases it.
func NewDir(path string) (*Dir, error) {
	if err := os.MkdirAll(path, dirMode); err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}

	return &Dir{root: root}, nil
}

// Put implements Store.
func (d *Dir) Put(ctx context.Context, name string, r io.Reader, size int64) error {
	if !validName(name) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if dir := filepath.Dir(name); dir != "." {
		if err := d.root.MkdirAll(dir, dirMode); err != nil {
			return err
		}
	}

	tmp := filepath.Join(filepath.Dir(name), ".tmp-"+uuid.New().String())

	f, err := d.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err == nil && n != size {
		err = fmt.Errorf("%w: wrote %d of %d bytes for %q", ErrShortObject, n, size, name)
	}

	if err == nil {
		err = f.Sync()
	}

	if closeErr := f.Close(); err == nil {
		err = closeErr
	}

	if err == nil {
		err = d.root.Rename(tmp, name)
	}

	if err != nil {
		if rmErr := d.root.Remove(tmp); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			err = errors.Join(err, rmErr)
		}

		return err
	}

	return nil
}

// validName accepts clean, local paths of a file: not ".", not ending in a separator, without ".." elements.
func validName(name string) bool {
	clean := filepath.Clean(name)

	return filepath.IsLocal(name) && clean == name && filepath.Base(clean) != "." &&
		!strings.HasSuffix(name, string(filepath.Separator)) && !strings.HasSuffix(name, "/")
}

// Close releases the directory.
func (d *Dir) Close() error { return d.root.Close() }
