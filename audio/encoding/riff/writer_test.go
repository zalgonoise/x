package riff_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// writeStream writes a RIFF file with one stream chunk holding body, and returns what closing the Writer returned.
func writeStream(t *testing.T, w io.Writer, body []byte) error {
	t.Helper()

	rw, err := riff.NewWriter(w, riff.WAVE)
	require.NoError(t, err)

	stream, err := rw.WriteStream(riff.ID("data"))
	require.NoError(t, err)

	_, err = stream.Write(body)
	require.NoError(t, err)

	return rw.Close()
}

// requireStream checks that b is a RIFF stream (unknown sizes) holding exactly body, with nothing appended.
func requireStream(t *testing.T, b, body []byte) {
	t.Helper()

	require.Len(t, b, 12+8+len(body), "nothing written after the data")
	require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(b[4:8]), "RIFF size")
	require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(b[16:20]), "data size")

	_, chunks := readAll(t, bytes.NewReader(b), riff.ID("data"))
	require.Equal(t, []chunk{{"data", body}}, chunks)
}

// A file opened with O_APPEND can seek, but every write goes to its end: back-patching would append the patches to the
// data. The file must stay a valid stream instead.
func TestWriterAppendModeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.riff")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)

	require.NoError(t, writeStream(t, f, []byte{1, 2, 3}))
	require.NoError(t, f.Close())

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	requireStream(t, b, []byte{1, 2, 3})
}

// /dev/null accepts writes and seeks without moving: writing a stream there must not fail.
func TestWriterDevNull(t *testing.T) {
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)

	defer f.Close()

	require.NoError(t, writeStream(t, f, []byte{1, 2, 3}))
}

var errTransient = errors.New("transient")

// flakyWriter fails once, after accepting part of the write that crosses failAt bytes.
type flakyWriter struct {
	failAt int
	out    []byte
	failed bool
}

func (f *flakyWriter) Write(p []byte) (int, error) {
	if !f.failed && len(f.out)+len(p) > f.failAt {
		f.failed = true
		n := f.failAt - len(f.out)
		f.out = append(f.out, p[:n]...)

		return n, errTransient
	}

	f.out = append(f.out, p...)

	return len(p), nil
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

func TestWriterErrorIsSticky(t *testing.T) {
	fw := &flakyWriter{failAt: 12 + 5} // fails inside the first chunk header

	w, err := riff.NewWriter(fw, riff.WAVE)
	require.NoError(t, err)
	require.ErrorIs(t, w.WriteChunk(riff.ID("LIST"), []byte("abcd")), errTransient)

	require.ErrorIs(t, w.WriteChunk(riff.ID("bext"), []byte{1, 2}), errTransient)

	_, err = w.Reserve(riff.ID("fact"), []byte{0, 0, 0, 0})
	require.ErrorIs(t, err, errTransient)

	_, err = w.WriteStream(riff.ID("data"))
	require.ErrorIs(t, err, errTransient)
	require.ErrorIs(t, w.Close(), errTransient)
	require.ErrorIs(t, w.Close(), errTransient, "a failed Close stays failed")
	require.Len(t, fw.out, 12+5, "nothing written after the failure")

	_, err = riff.NewWriter(shortWriter{}, riff.WAVE)
	require.ErrorIs(t, err, io.ErrShortWrite, "a short write without an error is an error")
}

func TestWriterStreamErrorIsSticky(t *testing.T) {
	fw := &flakyWriter{failAt: 12 + 8 + 3}

	w, err := riff.NewWriter(fw, riff.WAVE)
	require.NoError(t, err)

	stream, err := w.WriteStream(riff.ID("data"))
	require.NoError(t, err)

	n, err := stream.Write([]byte{1, 2, 3, 4, 5})
	require.ErrorIs(t, err, errTransient)
	require.Equal(t, 3, n)

	_, err = stream.Write([]byte{6})
	require.ErrorIs(t, err, errTransient)
	require.ErrorIs(t, w.Close(), errTransient)
}

var errNoRewind = errors.New("cannot seek backwards")

// forwardOnly is a WriteSeeker that reports its position but cannot rewind (an upload, a tape).
type forwardOnly struct{ out []byte }

func (f *forwardOnly) Write(p []byte) (int, error) {
	f.out = append(f.out, p...)

	return len(p), nil
}

func (f *forwardOnly) Seek(off int64, whence int) (int64, error) {
	if whence == io.SeekCurrent && off == 0 {
		return int64(len(f.out)), nil
	}

	return 0, errNoRewind
}

// When back-patching fails, nothing else is written: the data size stays unknown and no pad byte may follow the data.
func TestWriterFailedBackPatch(t *testing.T) {
	fo := &forwardOnly{}

	require.ErrorIs(t, writeStream(t, fo, []byte{1, 2, 3}), errNoRewind)
	requireStream(t, fo.out, []byte{1, 2, 3})
}

// The RIFF size bounds the walk once the stream chunk was seen: Next stops at its end and reads nothing beyond.
func TestReaderHonoursRIFFSize(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2}})
	trailing := []byte("\x00\x00\x00\x00\x00\x00\x00\x00RIFF") // NUL padding, then another file

	r := bytes.NewReader(append(bytes.Clone(file), trailing...))
	_, chunks := readAll(t, r, riff.ID("data"))

	require.Equal(t, []string{"fmt ", "data"}, []string{chunks[0].id, chunks[1].id})
	require.Len(t, chunks, 2)
	require.Equal(t, len(trailing), r.Len(), "nothing past the RIFF body is read")

	// before the stream chunk, a RIFF size that is too small is not trusted: it cannot hide the audio
	small := bytes.Clone(file)
	binary.LittleEndian.PutUint32(small[4:8], 4)
	_, chunks = readAll(t, bytes.NewReader(small), riff.ID("data"))
	require.Len(t, chunks, 2)
}

// A 0-size stream chunk is empty when a definite RIFF size declares chunks after it, and unbounded otherwise.
func TestReaderZeroSizeStreamChunk(t *testing.T) {
	zero := uint32(0)
	after := []testutil.Chunk{{ID: "LIST", Data: []byte("INFO")}}

	sized := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16, DataSize: &zero, After: after}.Encode(nil)
	rr, err := riff.NewReader(bytes.NewReader(sized), riff.ID("data"))
	require.NoError(t, err)

	_, _, err = rr.Next() // fmt
	require.NoError(t, err)

	_, body, err := rr.Next() // data
	require.NoError(t, err)
	require.False(t, rr.Unbounded())

	b, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Empty(t, b)

	h, _, err := rr.Next()
	require.NoError(t, err)
	require.Equal(t, riff.ID("LIST"), h.ID)

	// the RIFF size ends at the data header (a writer that never patched its sizes): the data runs to the end
	headerOnly := uint32(4 + 8 + 16 + 8) // "WAVE", the fmt chunk and the data chunk header
	live := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16, DataSize: &zero, RIFFSize: &headerOnly}.
		EncodeInts([][]int32{{1, 2}})
	rr, err = riff.NewReader(bytes.NewReader(live), riff.ID("data"))
	require.NoError(t, err)

	_, _, err = rr.Next()
	require.NoError(t, err)

	_, body, err = rr.Next()
	require.NoError(t, err)
	require.True(t, rr.Unbounded())

	b, err = io.ReadAll(body)
	require.NoError(t, err)
	require.Len(t, b, 4)
}

type stall struct{}

func (stall) Read([]byte) (int, error) { return 0, nil }

func TestReaderInputErrors(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2}})

	t.Run("NoProgress", func(t *testing.T) {
		rr, err := riff.NewReader(io.MultiReader(bytes.NewReader(file[:20]), stall{}))
		require.NoError(t, err)

		_, _, err = rr.Next()
		require.NoError(t, err)

		_, _, err = rr.Next() // skipping the fmt body stalls
		require.ErrorIs(t, err, io.ErrNoProgress)
	})

	t.Run("TransportErrorIsNotTruncation", func(t *testing.T) {
		rr, err := riff.NewReader(io.MultiReader(bytes.NewReader(file[:15]), errReader{io.ErrUnexpectedEOF}))
		require.NoError(t, err)

		_, _, err = rr.Next()
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.NotErrorIs(t, err, riff.ErrTruncated)
	})

	t.Run("IOErrorIsNotRIFF", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := riff.NewReader(errReader{boom})
		require.ErrorIs(t, err, boom)
		require.NotErrorIs(t, err, riff.ErrNotRIFF)

		_, err = riff.NewReader(nil)
		require.ErrorIs(t, err, riff.ErrNotRIFF)

		_, err = riff.NewWriter(nil, riff.WAVE)
		require.ErrorIs(t, err, riff.ErrNilWriter)
	})
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

// Round 2 of the Phase 2 review: the RIFF end bounds trailer headers and bodies, reader errors are sticky, and the
// Writer handles /dev/null, io.OffsetWriter, streams past 4 GiB and files another writer appends to.

func TestReaderBoundsTrailerByRIFFEnd(t *testing.T) {
	next := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{7}})

	t.Run("BodyPastTheEnd", func(t *testing.T) {
		// the LIST chunk declares 40 bytes, but the RIFF body holds only 4 of them
		file := testutil.WAV{
			Channels: 1, SampleRate: 8000, BitsPerSample: 16, After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFO")}},
		}.EncodeInts([][]int32{{1, 2}})
		binary.LittleEndian.PutUint32(file[len(file)-8:], 40)

		for _, read := range []bool{true, false} {
			r := bytes.NewReader(append(bytes.Clone(file), next...))
			rr, err := riff.NewReader(r, riff.ID("data"))
			require.NoError(t, err)

			var list []byte

			for {
				h, body, err := rr.Next()
				if errors.Is(err, io.EOF) {
					break
				}

				require.NoError(t, err)

				if h.ID == riff.ID("LIST") {
					require.Equal(t, uint32(40), h.Size, "the declared size is reported")

					if read {
						list, err = io.ReadAll(body)
						require.NoError(t, err)
					}
				}
			}

			if read {
				require.Equal(t, []byte("INFO"), list, "the body is cut at the end of the RIFF body")
			}

			require.Equal(t, len(next), r.Len(), "the next file is left alone (read=%v)", read)
		}
	})

	t.Run("HeaderPastTheEnd", func(t *testing.T) {
		// 3 bytes of a chunk header inside the RIFF body: reading the rest would take bytes of the next file
		file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2}})
		file = append(file, 'L', 'I', 'S')
		binary.LittleEndian.PutUint32(file[4:8], uint32(len(file)-8))

		r := bytes.NewReader(append(bytes.Clone(file), next...))
		rr, err := riff.NewReader(r, riff.ID("data"))
		require.NoError(t, err)

		var last error

		for last == nil {
			_, _, last = rr.Next()
		}

		require.ErrorIs(t, last, riff.ErrTruncated)
		require.Equal(t, len(next), r.Len())
	})
}

func TestReaderErrorIsSticky(t *testing.T) {
	file := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 16, Before: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOabcd")}},
	}.EncodeInts([][]int32{{1, 2}})

	// fail while skipping the LIST body; a retry must not parse the rest of it as a chunk header
	r := io.MultiReader(bytes.NewReader(file[:12+8+16+8+4]), errReader{errTransient}, bytes.NewReader(file[12+8+16+8+4:]))
	rr, err := riff.NewReader(r, riff.ID("data"))
	require.NoError(t, err)

	for range 2 {
		_, _, err = rr.Next()
		require.NoError(t, err)
	}

	_, _, err = rr.Next()
	require.ErrorIs(t, err, errTransient)

	_, _, err = rr.Next()
	require.ErrorIs(t, err, errTransient, "the walk cannot resume")
}

// stuckOffset is a WriteSeeker whose offset never moves, as Linux reports for /dev/null.
type stuckOffset struct{ n int }

func (s *stuckOffset) Write(p []byte) (int, error) {
	s.n += len(p)

	return len(p), nil
}

func (s *stuckOffset) Seek(int64, int) (int64, error) { return 0, nil }

func TestWriterOffsetDoesNotAdvance(t *testing.T) {
	s := &stuckOffset{}

	require.NoError(t, writeStream(t, s, []byte{1, 2, 3}))
	require.Equal(t, 12+8+3, s.n, "left as a stream: nothing patched, no pad byte")
}

func TestWriterOffsetWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.riff")

	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = f.WriteString("0123456789")
	require.NoError(t, err)

	// io.OffsetWriter writes in place through WriteAt but cannot seek to its end
	require.NoError(t, writeStream(t, io.NewOffsetWriter(f, 10), []byte{1, 2, 3}))
	require.NoError(t, f.Close())

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "0123456789", string(b[:10]))

	_, chunks := readAll(t, bytes.NewReader(b[10:]))
	require.Equal(t, []chunk{{"data", []byte{1, 2, 3}}}, chunks, "sizes patched")
	require.Equal(t, uint32(4+8+3+1), binary.LittleEndian.Uint32(b[14:18]))
}

// sparse is a seekable, patchable destination that only counts the bytes written after its header, so streams of 4 GiB
// can be written cheaply.
type sparse struct {
	head []byte
	size int64
	pos  int64
}

func (s *sparse) Write(p []byte) (int, error) {
	if room := 64 - len(s.head); room > 0 && s.pos == int64(len(s.head)) {
		s.head = append(s.head, p[:min(room, len(p))]...)
	}

	s.pos += int64(len(p))
	s.size = max(s.size, s.pos)

	return len(p), nil
}

func (s *sparse) WriteAt(p []byte, off int64) (int, error) {
	if off+int64(len(p)) <= int64(len(s.head)) {
		copy(s.head[off:], p)
	}

	s.size = max(s.size, off+int64(len(p)))

	return len(p), nil
}

func (s *sparse) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		s.pos = off
	case io.SeekCurrent:
		s.pos += off
	default:
		s.pos = s.size + off
	}

	return s.pos, nil
}

func TestWriterStreamOf4GiB(t *testing.T) {
	s := &sparse{}

	w, err := riff.NewWriter(s, riff.WAVE)
	require.NoError(t, err)

	stream, err := w.WriteStream(riff.ID("data"))
	require.NoError(t, err)

	chunk := make([]byte, 1<<20)
	for range 1 << 12 {
		_, err = stream.Write(chunk)
		require.NoError(t, err)
	}

	_, err = stream.Write([]byte{1}) // 4 GiB + 1: odd, and too large for the size field
	require.NoError(t, err)
	require.NoError(t, w.Close())

	require.Equal(t, int64(12+8+1<<32+1), s.size, "no pad byte after a size that reads as unknown")
	require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(s.head[16:20]))
}

// sharedFile is an *os.File another writer appends to (through its own O_APPEND descriptor) while Close probes it.
type sharedFile struct {
	*os.File

	other *os.File
	once  bool
}

func (s *sharedFile) WriteAt(p []byte, off int64) (int, error) {
	if len(p) > 0 && !s.once {
		s.once = true

		if _, err := s.other.WriteString("another writer's record\n"); err != nil {
			return 0, err
		}
	}

	return s.File.WriteAt(p, off)
}

// Close only ever truncates the 4 bytes of its own probe: growth it cannot account for is reported, not undone.
func TestWriterForeignGrowth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.riff")

	f, err := os.Create(path)
	require.NoError(t, err)

	defer f.Close()

	other, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	require.NoError(t, err)

	defer other.Close()

	require.ErrorIs(t, writeStream(t, &sharedFile{File: f, other: other}, []byte{1, 2, 3}), riff.ErrAppendOnly)

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, bytes.HasSuffix(b, []byte("another writer's record\n")), "the other writer's data is kept")
}
