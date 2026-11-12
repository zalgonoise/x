package iox_test

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/iox"
)

func TestSeekBuffer(t *testing.T) {
	var s iox.SeekBuffer

	n, err := s.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, 5, n)

	pos, err := s.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, int64(1), pos)

	_, err = s.Write([]byte("EL"))
	require.NoError(t, err)
	require.Equal(t, "hELlo", string(s.Bytes()))

	_, err = s.WriteAt([]byte("!"), 8) // past the end: the gap is zeroed
	require.NoError(t, err)
	require.Equal(t, []byte("hELlo\x00\x00\x00!"), s.Bytes())
	require.Equal(t, 9, s.Len())

	pos, err = s.Seek(0, io.SeekEnd)
	require.NoError(t, err)
	require.Equal(t, int64(9), pos)

	pos, err = s.Seek(-2, io.SeekCurrent)
	require.NoError(t, err)
	require.Equal(t, int64(7), pos)

	_, err = s.Seek(-10, io.SeekEnd)
	require.ErrorIs(t, err, iox.ErrNegativeOffset)

	_, err = s.WriteAt([]byte("x"), -1)
	require.ErrorIs(t, err, iox.ErrNegativeOffset)

	_, err = s.Seek(0, 42)
	require.ErrorIs(t, err, iox.ErrWhence)

	_, err = s.WriteAt([]byte("x"), math.MaxInt64)
	require.ErrorIs(t, err, iox.ErrTooLarge, "an offset that overflows is an error, not a panic")

	_, err = s.Seek(math.MaxInt64-1, io.SeekStart)
	require.NoError(t, err)

	_, err = s.Write([]byte("xy"))
	require.ErrorIs(t, err, iox.ErrTooLarge)

	s.Reset()
	require.Zero(t, s.Len())

	_, err = s.Write([]byte("ab"))
	require.NoError(t, err)
	require.Equal(t, "ab", string(s.Bytes()), "a reused buffer holds no stale bytes")
}
