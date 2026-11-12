package data

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/testutil"
)

func TestReadFrames(t *testing.T) {
	src := make([]byte, 1000)
	for i := range src {
		src[i] = byte(i)
	}

	collect := func(t *testing.T, r io.Reader, align int, limit int64) ([]byte, int64) {
		t.Helper()

		var got []byte

		n, err := readFrames(r, align, 64, limit, func(b []byte) error {
			require.Zero(t, len(b)%align, "every call holds whole frames")
			got = append(got, b...)

			return nil
		})
		require.NoError(t, err)

		return got, n
	}

	for _, sizes := range testutil.ReadSizes {
		for _, align := range []int{1, 2, 3, 4, 6, 8} {
			got, n := collect(t, testutil.ChunkReader(bytes.NewReader(src), sizes...), align, 0)
			whole := len(src) - len(src)%align

			require.Equal(t, int64(len(src)), n, "every byte is counted")
			require.Equal(t, src[:whole], got, "sizes %v align %d", sizes, align)
		}
	}

	t.Run("DataWithEOF", func(t *testing.T) {
		got, _ := collect(t, iotest.DataErrReader(bytes.NewReader(src)), 4, 0)
		require.Equal(t, src, got)
	})

	t.Run("OneByteReader", func(t *testing.T) {
		got, _ := collect(t, iotest.OneByteReader(bytes.NewReader(src)), 6, 0)
		require.Equal(t, src[:996], got)
	})

	t.Run("Limit", func(t *testing.T) {
		r := bytes.NewReader(src)
		got, n := collect(t, testutil.ChunkReader(r, 7), 2, 101)
		require.Equal(t, int64(101), n)
		require.Equal(t, src[:100], got, "the partial frame at the limit is dropped")
		require.Equal(t, len(src)-101, r.Len(), "nothing past the limit is consumed")
	})

	t.Run("ReadError", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := readFrames(iotest.ErrReader(boom), 2, 64, 0, func([]byte) error { return nil })
		require.ErrorIs(t, err, boom)
	})

	t.Run("FnError", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := readFrames(bytes.NewReader(src), 2, 64, 0, func([]byte) error { return boom })
		require.ErrorIs(t, err, boom)
	})

	t.Run("NoProgress", func(t *testing.T) {
		_, err := readFrames(zeroReader{}, 2, 64, 0, func([]byte) error { return nil })
		require.ErrorIs(t, err, io.ErrNoProgress)
	})

	t.Run("NilReader", func(t *testing.T) {
		_, err := readFrames(nil, 2, 64, 0, func([]byte) error { return nil })
		require.ErrorIs(t, err, ErrNilReader)
	})
}

type zeroReader struct{}

func (zeroReader) Read([]byte) (int, error) { return 0, nil }

func TestDeclaredSize(t *testing.T) {
	require.Zero(t, declaredSize(nil))
	require.Zero(t, declaredSize(&Header{Subchunk2Size: 0}))
	require.Zero(t, declaredSize(&Header{Subchunk2Size: math.MaxUint32}))
	require.Equal(t, int64(37393), declaredSize(&Header{Subchunk2Size: 37393}))
}

func TestChunkReadFrom(t *testing.T) {
	payload := make([]byte, 0, 2*2*100)
	for i := range 2 * 100 {
		payload = append(payload, byte(i), byte(i>>8)) // 16-bit samples, stereo frames of 4 bytes
	}

	for _, sizes := range testutil.ReadSizes {
		chunk := NewPCMChunk(16, &Header{Subchunk2ID: DataID, Subchunk2Size: uint32(len(payload))})
		chunk.SetBlockAlign(4)

		r := bytes.NewReader(append(bytes.Clone(payload), []byte("LIST")...))
		n, err := chunk.ReadFrom(testutil.ChunkReader(r, sizes...))
		require.NoError(t, err)
		require.Equal(t, int64(len(payload)), n)
		require.Equal(t, payload, chunk.Bytes(), "sizes %v", sizes)
		require.Equal(t, 4, r.Len(), "the next chunk is left unread")
	}

	t.Run("OddSizeSkipsPad", func(t *testing.T) {
		chunk := NewPCMChunk(8, &Header{Subchunk2ID: DataID, Subchunk2Size: 3})
		r := bytes.NewReader([]byte{1, 2, 3, 0, 'n', 'e', 'x', 't'})

		n, err := chunk.ReadFrom(r)
		require.NoError(t, err)
		require.Equal(t, int64(4), n)
		require.Len(t, chunk.Data, 3, "the pad byte is not a sample")
		require.Equal(t, 4, r.Len())
	})

	t.Run("LiveSizeReadsToEOF", func(t *testing.T) {
		for _, size := range []uint32{0, math.MaxUint32} {
			chunk := NewPCMChunk(16, &Header{Subchunk2ID: DataID, Subchunk2Size: size})
			_, err := chunk.ReadFrom(iotest.HalfReader(bytes.NewReader(payload)))
			require.NoError(t, err)
			require.Equal(t, payload, chunk.Bytes())
		}
	})
}

func TestRingReadFromWholeFrames(t *testing.T) {
	payload := make([]byte, 6*500) // 24-bit stereo frames
	for i := range payload {
		payload[i] = byte(i * 7)
	}

	for _, sizes := range testutil.ReadSizes {
		var got []float64

		ring := NewPCMRing(24, &Header{Subchunk2ID: DataID}, 600, func(f []float64) error {
			require.Zero(t, len(f)%2, "the processor always receives whole stereo frames")
			got = append(got, f...)

			return nil
		})
		ring.SetBlockAlign(6)

		_, err := ring.ReadFrom(testutil.ChunkReader(iotest.DataErrReader(bytes.NewReader(payload)), sizes...))
		require.NoError(t, err)

		want := ring.Converter.Parse(payload)
		require.Equal(t, want, got, "sizes %v", sizes)
	}
}

func TestJunkReadFrom(t *testing.T) {
	t.Run("ShortReads", func(t *testing.T) {
		junk := NewJunk(&Header{Subchunk2ID: JunkID, Subchunk2Size: 5})
		r := bytes.NewReader([]byte("abcde\x00data"))

		n, err := junk.ReadFrom(iotest.OneByteReader(r))
		require.NoError(t, err)
		require.Equal(t, int64(6), n, "body and pad byte")
		require.Equal(t, []byte("abcde"), junk.Data)
		require.Equal(t, 4, r.Len())
	})

	t.Run("LyingHeaderDoesNotAllocate", func(t *testing.T) {
		junk := NewJunk(&Header{Subchunk2ID: JunkID, Subchunk2Size: 128 << 20})

		_, err := junk.ReadFrom(bytes.NewReader(make([]byte, 100)))
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.Less(t, cap(junk.Data), 1<<20)
	})

	t.Run("AfterResetNoPanic", func(t *testing.T) {
		junk := NewJunk(&Header{Subchunk2ID: JunkID, Subchunk2Size: 4})
		junk.Reset()

		_, err := junk.ReadFrom(bytes.NewReader([]byte("abcd")))
		require.NoError(t, err)
		require.Equal(t, []byte("abcd"), junk.Data)
	})
}

func TestHeaderReadFromShortReads(t *testing.T) {
	h := new(Header)

	_, err := h.ReadFrom(iotest.OneByteReader(bytes.NewReader([]byte("LIST\x04\x00\x00\x00"))))
	require.NoError(t, err)
	require.Equal(t, "LIST", string(h.Subchunk2ID[:]))
	require.Equal(t, uint32(4), h.Subchunk2Size)

	_, err = new(Header).ReadFrom(bytes.NewReader([]byte("dat")))
	require.ErrorIs(t, err, ErrShortBuffer)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	_, err = new(Header).ReadFrom(nil)
	require.ErrorIs(t, err, ErrNilReader)
}

func TestChunkSize(t *testing.T) {
	for _, test := range []struct {
		n    int
		want uint32
	}{
		{n: -1, want: 0},
		{n: 0, want: 0},
		{n: 7, want: 7},
		{n: math.MaxUint32, want: math.MaxUint32},
		{n: math.MaxUint32 + 1, want: math.MaxUint32},
		{n: math.MaxInt, want: math.MaxUint32},
	} {
		require.Equal(t, test.want, chunkSize(test.n), "n=%d", test.n)
	}
}
