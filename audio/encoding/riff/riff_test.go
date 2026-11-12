package riff_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

type chunk struct {
	id   string
	body []byte
}

func readAll(t *testing.T, r io.Reader, unbounded ...riff.FourCC) (riff.FourCC, []chunk) {
	t.Helper()

	rr, err := riff.NewReader(r, unbounded...)
	require.NoError(t, err)

	var out []chunk

	for {
		h, body, err := rr.Next()
		if errors.Is(err, io.EOF) {
			return rr.Form(), out
		}

		require.NoError(t, err)

		b, err := io.ReadAll(body)
		require.NoError(t, err)

		out = append(out, chunk{id: h.ID.String(), body: b})
	}
}

func TestReaderWalksBuilderFiles(t *testing.T) {
	w := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 8,
		Before: []testutil.Chunk{{ID: "LIST", Data: []byte("odd")}, {ID: "JUNK", Data: make([]byte, 4)}},
		After:  []testutil.Chunk{{ID: "id3 ", Data: []byte{1, 2, 3}}},
	}
	file := w.EncodeInts([][]int32{{1, 2, 3, 4, 5}}) // odd data chunk too

	for _, r := range []io.Reader{
		bytes.NewReader(file),
		iotest.OneByteReader(bytes.NewReader(file)),
		iotest.DataErrReader(testutil.ChunkReader(bytes.NewReader(file), 3)),
	} {
		form, chunks := readAll(t, r)
		require.Equal(t, riff.WAVE, form)

		ids := make([]string, 0, len(chunks))
		for _, c := range chunks {
			ids = append(ids, c.id)
		}

		require.Equal(t, []string{"fmt ", "LIST", "JUNK", "data", "id3 "}, ids)
		require.Equal(t, []byte("odd"), chunks[1].body)
		require.Equal(t, []byte{129, 130, 131, 132, 133}, chunks[3].body)
		require.Equal(t, []byte{1, 2, 3}, chunks[4].body)
	}
}

func TestReaderSkipsUnreadBodies(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_mono_8bit_176400hz.wav")

	rr, err := riff.NewReader(bytes.NewReader(file))
	require.NoError(t, err)

	var ids []string

	for {
		h, _, err := rr.Next() // never read the bodies
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		ids = append(ids, h.ID.String())
	}

	require.Equal(t, []string{"fmt ", "junk", "data"}, ids, "odd data chunk + pad byte handled")
}

func TestReaderLiveSize(t *testing.T) {
	size := uint32(riff.UnknownSize)
	file := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 16, DataSize: &size,
		After: []testutil.Chunk{{ID: "LIST", Data: []byte("tail")}}, // part of the stream once the size is unknown
	}.EncodeInts([][]int32{{1, 2}})

	_, chunks := readAll(t, bytes.NewReader(file), riff.ID("data"))
	require.Len(t, chunks, 2)
	require.Equal(t, "data", chunks[1].id)
	require.Len(t, chunks[1].body, 4+8+4, "the unbounded data chunk runs to the end of the input")
}

func TestReaderErrors(t *testing.T) {
	_, err := riff.NewReader(bytes.NewReader([]byte("RIFX\x00\x00\x00\x00WAVE")))
	require.ErrorIs(t, err, riff.ErrNotRIFF)

	_, err = riff.NewReader(bytes.NewReader([]byte("RIF")))
	require.ErrorIs(t, err, riff.ErrNotRIFF)

	rr, err := riff.NewReader(bytes.NewReader([]byte("RIFF\x10\x00\x00\x00WAVE\x00\x01\x02\x03\x04\x00\x00\x00")))
	require.NoError(t, err)

	_, _, err = rr.Next()
	require.ErrorIs(t, err, riff.ErrBadID)

	rr, err = riff.NewReader(bytes.NewReader([]byte("RIFF\x10\x00\x00\x00WAVEfmt ")))
	require.NoError(t, err)

	_, _, err = rr.Next()
	require.ErrorIs(t, err, riff.ErrTruncated)
}

func TestWriterRoundTrip(t *testing.T) {
	var buf bytes.Buffer

	w, err := riff.NewWriter(&buf, riff.WAVE)
	require.NoError(t, err)
	require.NoError(t, w.WriteChunk(riff.ID("fmt"), bytes.Repeat([]byte{1}, 16)))
	require.NoError(t, w.WriteChunk(riff.ID("LIST"), []byte("odd")))
	require.NoError(t, w.WriteChunk(riff.ID("data"), []byte{1, 2, 3, 4}))
	require.NoError(t, w.Close())

	_, chunks := readAll(t, bytes.NewReader(buf.Bytes()))
	require.Equal(t, []chunk{{"fmt ", bytes.Repeat([]byte{1}, 16)}, {"LIST", []byte("odd")}, {"data", []byte{1, 2, 3, 4}}}, chunks)
	require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(buf.Bytes()[4:8]),
		"a plain io.Writer cannot be back-patched")
}

func TestWriterBackPatchesSeekable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.wav")

	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = f.WriteString("prefix") // the RIFF file does not start at offset 0
	require.NoError(t, err)

	w, err := riff.NewWriter(f, riff.WAVE)
	require.NoError(t, err)
	require.NoError(t, w.WriteChunk(riff.ID("fmt"), make([]byte, 16)))

	stream, err := w.WriteStream(riff.ID("data"))
	require.NoError(t, err)

	_, err = stream.Write([]byte{1, 2, 3})
	require.NoError(t, err)
	require.ErrorIs(t, w.WriteChunk(riff.ID("LIST"), nil), riff.ErrClosed, "nothing follows a stream chunk")
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	b, err := os.ReadFile(path)
	require.NoError(t, err)

	file := b[len("prefix"):]
	require.Equal(t, uint32(len(file)-8), binary.LittleEndian.Uint32(file[4:8]), "RIFF size back-patched")

	_, chunks := readAll(t, bytes.NewReader(file))
	require.Equal(t, "data", chunks[1].id)
	require.Equal(t, []byte{1, 2, 3}, chunks[1].body, "stream size back-patched to 3, pad byte appended")
	require.Equal(t, 12+8+16+8+3+1, len(file))
}

func TestWriterReserve(t *testing.T) {
	write := func(t *testing.T, w io.Writer) {
		t.Helper()

		rw, err := riff.NewWriter(w, riff.WAVE)
		require.NoError(t, err)

		r, err := rw.Reserve(riff.ID("fact"), []byte{0xFF, 0xFF, 0xFF, 0xFF})
		require.NoError(t, err)

		odd, err := rw.Reserve(riff.ID("odd"), []byte{9, 9, 9})
		require.NoError(t, err)

		stream, err := rw.WriteStream(riff.ID("data"))
		require.NoError(t, err)

		_, err = stream.Write([]byte{1, 2})
		require.NoError(t, err)

		require.ErrorIs(t, r.Set([]byte{1}), riff.ErrSize)
		require.NoError(t, r.Set([]byte{1, 0, 0, 0}))
		require.NoError(t, odd.Set([]byte{7, 8, 9}))
		require.NoError(t, rw.Close())
	}

	t.Run("Seekable", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "out.riff"))
		require.NoError(t, err)

		_, err = f.WriteString("prefix")
		require.NoError(t, err)

		write(t, f)

		b, err := os.ReadFile(f.Name())
		require.NoError(t, err)
		require.NoError(t, f.Close())

		_, chunks := readAll(t, bytes.NewReader(b[len("prefix"):]))
		require.Equal(t, []chunk{{"fact", []byte{1, 0, 0, 0}}, {"odd ", []byte{7, 8, 9}}, {"data", []byte{1, 2}}}, chunks)
	})

	t.Run("Stream", func(t *testing.T) {
		var buf bytes.Buffer

		write(t, &buf)

		_, chunks := readAll(t, bytes.NewReader(buf.Bytes()), riff.ID("data"))
		require.Equal(t, []chunk{{"fact", []byte{0xFF, 0xFF, 0xFF, 0xFF}}, {"odd ", []byte{9, 9, 9}}, {"data", []byte{1, 2}}},
			chunks, "without seeking, reserved chunks keep their first body")
	})
}

func TestWriterStreamIsNotPadded(t *testing.T) {
	write := func(t *testing.T, w io.Writer) {
		t.Helper()

		rw, err := riff.NewWriter(w, riff.WAVE)
		require.NoError(t, err)

		stream, err := rw.WriteStream(riff.ID("data"))
		require.NoError(t, err)

		_, err = stream.Write([]byte{1, 2, 3})
		require.NoError(t, err)
		require.NoError(t, rw.Close())
	}

	t.Run("Writer", func(t *testing.T) {
		var buf bytes.Buffer

		write(t, &buf)

		_, chunks := readAll(t, bytes.NewReader(buf.Bytes()), riff.ID("data"))
		require.Equal(t, []chunk{{"data", []byte{1, 2, 3}}}, chunks, "a pad byte would be read as data")
	})

	t.Run("Pipe", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)

		type result struct {
			b   []byte
			err error
		}

		done := make(chan result)

		go func() {
			b, err := io.ReadAll(r)
			done <- result{b, err}
		}()

		write(t, w) // *os.File implements io.WriteSeeker, but a pipe cannot seek
		require.NoError(t, w.Close())

		res := <-done
		require.NoError(t, res.err)
		require.NoError(t, r.Close())

		b := res.b

		_, chunks := readAll(t, bytes.NewReader(b), riff.ID("data"))
		require.Equal(t, []chunk{{"data", []byte{1, 2, 3}}}, chunks)
		require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(b[4:8]))
	})
}

func TestFourCC(t *testing.T) {
	require.Equal(t, "fmt ", riff.ID("fmt").String())
	require.True(t, riff.ID("LIST").Valid())
	require.False(t, riff.FourCC{0, 'a', 'b', 'c'}.Valid())
}
