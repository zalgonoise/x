//go:build unix

package riff_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// A descriptor opened with O_APPEND outside os.OpenFile (stdout redirected with >>) looks like a regular *os.File:
// WriteAt works, but every write still lands at the end. The Writer detects it, undoes its probe and leaves a stream.
func TestWriterAppendModeDescriptor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.riff")

	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND, 0o600)
	require.NoError(t, err)

	f := os.NewFile(uintptr(fd), path)

	require.NoError(t, writeStream(t, f, []byte{1, 2, 3}))
	require.NoError(t, f.Close())

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	requireStream(t, b, []byte{1, 2, 3})
}
