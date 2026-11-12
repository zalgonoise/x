package wav_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// Regression tests for the 2026-10-20 review of the Phase 2 Decoder and Encoder (CLAUDE.md §10).

// rawWAV builds a RIFF/WAVE file from a raw fmt body and a data body, with a correct RIFF size.
func rawWAV(fmtBody, data []byte) []byte {
	le := binary.LittleEndian
	chunk := func(dst []byte, id string, body []byte) []byte {
		dst = append(dst, id...)
		dst = le.AppendUint32(dst, uint32(len(body)))
		dst = append(dst, body...)

		if len(body)%2 == 1 {
			dst = append(dst, 0)
		}

		return dst
	}

	body := chunk(chunk([]byte("WAVE"), "fmt ", fmtBody), "data", data)

	return append(append([]byte("RIFF"), le.AppendUint32(nil, uint32(len(body)))...), body...)
}

// fmtEx builds a WAVEFORMATEX body with the given extra bytes after cbSize.
func fmtEx(tag, channels uint16, rate, byteRate uint32, blockAlign, bits uint16, extra ...byte) []byte {
	le := binary.LittleEndian
	b := le.AppendUint16(nil, tag)
	b = le.AppendUint16(b, channels)
	b = le.AppendUint32(b, rate)
	b = le.AppendUint32(b, byteRate)
	b = le.AppendUint16(b, blockAlign)
	b = le.AppendUint16(b, bits)
	b = le.AppendUint16(b, uint16(len(extra)))

	return append(b, extra...)
}

// Compressed formats follow their own block-align and bit rules (WAVEFORMATEX docs: the PCM formula applies to PCM):
// they are well-formed, so unsupported rather than invalid.
func TestDecoderCompressedFormatsAreUnsupported(t *testing.T) {
	for name, fmtBody := range map[string][]byte{
		"IMAADPCM": fmtEx(0x0011, 1, 22050, 11100, 512, 4, 0xF9, 0x03),
		"MP3":      fmtEx(0x0055, 2, 44100, 16000, 1, 0, 1, 0, 2, 0, 0, 0, 0xA1, 0x01, 1, 0, 0x71, 0x05),
		"GSM610":   fmtEx(0x0031, 1, 8000, 1625, 65, 0, 0x40, 0x01),
	} {
		_, err := wav.NewDecoder(bytes.NewReader(rawWAV(fmtBody, make([]byte, 512))))
		require.ErrorIs(t, err, audio.ErrUnsupportedFormat, name)
		require.NotErrorIs(t, err, audio.ErrInvalidFormat, name)
	}
}

// A 0-size data chunk in a file whose RIFF size covers chunks after it is empty, not a live stream.
func TestDecoderEmptyDataChunk(t *testing.T) {
	file := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 16,
		After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOISFT\x06\x00\x00\x00Lavf6\x00")}},
	}.Encode(nil)

	d, err := wav.NewDecoder(bytes.NewReader(file))
	require.NoError(t, err)

	frames, ok := d.Frames()
	require.True(t, ok)
	require.Zero(t, frames)

	buf, err := audio.ReadAll(d)
	require.NoError(t, err)
	require.Empty(t, buf.Samples, "the LIST chunk is not audio")
	require.NoError(t, d.ReadTrailer())
	require.Len(t, d.Chunks(), 1)
}

// Concatenated files: ReadTrailer stops at the end of the first RIFF body, so the second one can be decoded next.
func TestDecoderConcatenatedFiles(t *testing.T) {
	one := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8}.EncodeInts([][]int32{{1, 2, 3}}) // odd: padded
	two := testutil.WAV{Channels: 2, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{3}, {4}})
	r := bytes.NewReader(append(bytes.Clone(one), two...))

	for _, want := range [][]float64{scaled(8, 1, 2, 3), scaled(16, 3, 4)} {
		d, err := wav.NewDecoder(r)
		require.NoError(t, err)

		got, err := readAllSamples(t, d, 2)
		require.NoError(t, err)
		requireSameSamples(t, want, got)
		require.NoError(t, d.ReadTrailer())
		require.Empty(t, d.Chunks())
	}

	require.Zero(t, r.Len())
}

func allocated(fn func()) uint64 {
	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)

	return after.TotalAlloc - before.TotalAlloc
}

// A header declaring many channels must not make decoding allocate out of proportion to the data present.
func TestDecoderHostileChannelCounts(t *testing.T) {
	for name, file := range map[string][]byte{
		// 8191 channels of float64 (block align 65528), 2 data bytes
		"Float64": rawWAV(fmtEx(0x0003, 8191, 8000, 8000*65528, 65528, 64), []byte{0, 0}),
		// 65535 channels of 8 bits (block align 65535), no data
		"PCM8": testutil.WAV{Channels: 65535, SampleRate: 8000, BitsPerSample: 8}.Encode(nil),
	} {
		var samples int

		n := allocated(func() {
			sr, _, err := audio.Decode(bytes.NewReader(file))
			require.NoError(t, err, name)

			buf, err := audio.ReadAll(sr)
			require.NoError(t, err, name)

			samples = len(buf.Samples)
		})

		require.Zero(t, samples, name)
		require.Less(t, n, uint64(16<<20), "%s: allocated %d MiB for a %d-byte file", name, n>>20, len(file))
	}
}

func TestDecoderReadTrailerIOErrorIsSticky(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2, 3, 4}})
	boom := errors.New("boom")

	d, err := wav.NewDecoder(io.MultiReader(bytes.NewReader(file[:len(file)-2]), errReader{boom}))
	require.NoError(t, err)

	require.ErrorIs(t, d.ReadTrailer(), boom, "skipping the unread audio fails")
	require.ErrorIs(t, d.ReadTrailer(), boom)

	_, err = d.ReadSamples(make([]float64, 4))
	require.ErrorIs(t, err, boom, "not a clean io.EOF")
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

type stallReader struct{}

func (stallReader) Read([]byte) (int, error) { return 0, nil }

func TestDecoderNoProgress(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2, 3, 4}})

	d, err := wav.NewDecoder(io.MultiReader(bytes.NewReader(file[:44]), stallReader{}))
	require.NoError(t, err)

	_, err = d.ReadSamples(make([]float64, 4))
	require.ErrorIs(t, err, io.ErrNoProgress)

	_, err = wav.NewDecoder(io.MultiReader(bytes.NewReader(file[:30]), stallReader{}))
	require.ErrorIs(t, err, io.ErrNoProgress, "in the fmt chunk")
}

// The metadata limit bounds what the listed chunks keep alive, overhead included.
func TestDecoderMetadataLimitBoundsMemory(t *testing.T) {
	const limit = 1 << 20

	before := make([]testutil.Chunk, limit/8)
	for i := range before {
		before[i] = testutil.Chunk{ID: "JUNK"}
	}

	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16, Before: before}.EncodeInts([][]int32{{1}})

	var m0, m1 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m0)

	d, err := wav.NewDecoder(bytes.NewReader(file), wav.WithMetadataLimit(limit))
	require.NoError(t, err)

	runtime.GC()
	runtime.ReadMemStats(&m1)

	retained := int64(m1.HeapAlloc) - int64(m0.HeapAlloc)

	require.Len(t, d.Chunks(), limit/wav.ChunkOverhead)
	require.LessOrEqual(t, retained, int64(2*limit), "retained %d KiB", retained>>10)
	runtime.KeepAlive(d)
	runtime.KeepAlive(file)
}

// A dropped connection is not the end of the stream: net/http reports io.ErrUnexpectedEOF, and io.EOF is reserved for
// the end of the audio.
func TestDecoderDroppedConnection(t *testing.T) {
	f := fixtures.All(t, func(m fixtures.Meta) bool { return m.Channels == 2 && m.BitsPerSample == 16 })[0]

	for name, opts := range map[string][]testutil.ServeOption{
		"LiveStream":     {testutil.Endless(f.DataSize), testutil.DisconnectAfter(20000)},
		"BoundedChunked": {testutil.Chunked(512), testutil.DisconnectAfter(len(f.Data) / 2)},
	} {
		t.Run(name, func(t *testing.T) {
			srv := testutil.NewWAVServer(t, f.Data, opts...)

			res, err := http.Get(srv.URL)
			require.NoError(t, err)

			defer res.Body.Close()

			d, err := wav.NewDecoder(res.Body)
			require.NoError(t, err)

			dst := make([]float64, 256)

			for err == nil {
				_, err = d.ReadSamples(dst)
			}

			require.NotErrorIs(t, err, io.EOF)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
	}
}

func TestEncoderFullValidBitsIsNotExtensible(t *testing.T) {
	f := audio.Format{SampleRate: 44100, Channels: 2, BitDepth: 16, ValidBits: 16, Encoding: audio.EncodingPCM}

	var buf bytes.Buffer

	e, err := wav.NewEncoder(&buf, f)
	require.NoError(t, err)
	require.NoError(t, e.Close())

	b := buf.Bytes()
	require.Equal(t, uint32(16), binary.LittleEndian.Uint32(b[16:20]), "fmt chunk size")
	require.Equal(t, wav.PCMFormat, binary.LittleEndian.Uint16(b[20:22]), "wFormatTag")
}

func TestEncoderNilWriter(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

	e, err := wav.NewEncoder(nil, f)
	require.Error(t, err)
	require.Nil(t, e)
}

// A file opened with O_APPEND cannot be back-patched: the output stays a valid live stream with the same samples.
func TestEncoderAppendModeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.wav")

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)

	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingPCM}
	samples := scaled(8, 1, -2, 3) // odd: a pad byte would be read as a sample

	e, err := wav.NewEncoder(out, f)
	require.NoError(t, err)

	_, err = e.WriteSamples(samples)
	require.NoError(t, err)
	require.NoError(t, e.Close())
	require.NoError(t, out.Close())

	b, err := os.ReadFile(path)
	require.NoError(t, err)

	d, got := decodeBytes(t, b)
	requireSameSamples(t, samples, got)

	_, ok := d.Frames()
	require.False(t, ok, "sizes left unknown")
}

// Round 2: stray fmt and data chunks after the audio are skipped, so Chunks stays writable metadata.
func TestDecoderTrailerSkipsStrayChunks(t *testing.T) {
	list := testutil.Chunk{ID: "LIST", Data: []byte("INFO")}
	file := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 16,
		After: []testutil.Chunk{{ID: "fmt ", Data: make([]byte, 16)}, list, {ID: "data", Data: []byte{1, 2}}},
	}.EncodeInts([][]int32{{1, 2}})

	d, got := decodeBytes(t, file)
	requireSameSamples(t, scaled(16, 1, 2), got)
	require.NoError(t, d.ReadTrailer())
	require.Len(t, d.Chunks(), 1)
	require.Equal(t, riff.ID("LIST"), d.Chunks()[0].ID)

	_, err := wav.NewEncoder(io.Discard, d.Format(), wav.WithChunks(d.Chunks()...))
	require.NoError(t, err, "the trailer transcodes")
}

// Round 2: wValidBitsPerSample may not exceed the declared wBitsPerSample, even when the container rounds it up.
func TestParseFormatValidBitsAboveDeclared(t *testing.T) {
	for _, tc := range []struct{ bits, valid uint16 }{{12, 16}, {20, 24}, {16, 17}} {
		container := (tc.bits + 7) / 8
		b := testutil.WAV{
			Format: testutil.FormatExtensible, Channels: 1, SampleRate: 8000, BitsPerSample: container * 8, ValidBits: tc.valid,
		}.Encode(nil)
		binary.LittleEndian.PutUint16(b[fmtOff+14:], tc.bits) // declared wBitsPerSample below the container

		_, err := wav.ParseFormat(b[fmtOff : fmtOff+40])
		require.ErrorIs(t, err, audio.ErrInvalidFormat, "%d valid bits declared in %d", tc.valid, tc.bits)
	}

	b := testutil.WAV{Format: testutil.FormatExtensible, Channels: 1, SampleRate: 8000, BitsPerSample: 24, ValidBits: 20}.
		Encode(nil)
	binary.LittleEndian.PutUint16(b[fmtOff+14:], 20)

	f, err := wav.ParseFormat(b[fmtOff : fmtOff+40])
	require.NoError(t, err, "20 valid bits in a declared 20-bit, 24-bit container")
	require.Equal(t, 24, f.BitDepth)
	require.Equal(t, 20, f.ValidBits)
}
