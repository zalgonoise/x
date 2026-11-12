package iox_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/iox"
)

type stall struct{}

func (stall) Read([]byte) (int, error) { return 0, nil }

type script []struct {
	b   []byte
	err error
}

func (s *script) Read(p []byte) (int, error) {
	if len(*s) == 0 {
		return 0, io.EOF
	}

	step := (*s)[0]
	*s = (*s)[1:]

	return copy(p, step.b), step.err
}

func TestReadAtLeast(t *testing.T) {
	buf := make([]byte, 8)

	n, err := iox.ReadAtLeast(iotest.OneByteReader(bytes.NewReader([]byte("abcdefgh"))), buf, 5)
	require.NoError(t, err)
	require.Equal(t, 5, n)

	n, err = iox.ReadAtLeast(bytes.NewReader([]byte("abc")), buf, 5)
	require.ErrorIs(t, err, io.EOF, "a short input is io.EOF, not io.ErrUnexpectedEOF")
	require.NotErrorIs(t, err, io.ErrUnexpectedEOF)
	require.Equal(t, 3, n)

	n, err = iox.ReadAtLeast(iotest.DataErrReader(bytes.NewReader([]byte("abcde"))), buf, 5)
	require.NoError(t, err, "io.EOF returned with the last bytes is dropped once min is reached")
	require.Equal(t, 5, n)

	dropped := &script{{b: []byte("ab")}, {err: io.ErrUnexpectedEOF}}
	n, err = iox.ReadAtLeast(dropped, buf, 5)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF, "the reader's own io.ErrUnexpectedEOF is passed through")
	require.Equal(t, 2, n)

	boom := errors.New("boom")
	n, err = iox.ReadAtLeast(&script{{b: []byte("abcdef"), err: boom}}, buf, 5)
	require.ErrorIs(t, err, boom, "other errors are returned with the bytes")
	require.Equal(t, 6, n)

	n, err = iox.ReadAtLeast(io.MultiReader(bytes.NewReader([]byte("ab")), stall{}), buf, 5)
	require.ErrorIs(t, err, io.ErrNoProgress)
	require.Equal(t, 2, n)
}

func TestProgress(t *testing.T) {
	_, err := io.Copy(io.Discard, iox.Progress(io.MultiReader(bytes.NewReader([]byte("abc")), stall{})))
	require.ErrorIs(t, err, io.ErrNoProgress)

	p := iox.Progress(bytes.NewReader([]byte("abc")))
	require.Same(t, p, iox.Progress(p), "not wrapped twice")

	b, err := io.ReadAll(iox.Progress(iotest.HalfReader(bytes.NewReader([]byte("abcdef")))))
	require.NoError(t, err)
	require.Equal(t, []byte("abcdef"), b)
}
