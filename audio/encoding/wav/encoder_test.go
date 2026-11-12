package wav_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

var _ audio.SampleWriter = (*wav.Encoder)(nil)

// encodeFile encodes samples into a file under t.TempDir() (an io.WriteSeeker, so sizes are back-patched) and returns its
// contents.
func encodeFile(t *testing.T, f audio.Format, samples []float64, opts ...wav.EncoderOption) []byte {
	t.Helper()

	out, err := os.Create(filepath.Join(t.TempDir(), "out.wav"))
	require.NoError(t, err)

	defer out.Close()

	e, err := wav.NewEncoder(out, f, opts...)
	require.NoError(t, err)
	require.Equal(t, f, e.Format())

	n, err := e.WriteSamples(samples)
	require.NoError(t, err)
	require.Equal(t, len(samples), n)
	require.Equal(t, int64(len(samples)/f.Channels), e.Frames())
	require.NoError(t, e.Close())
	require.NoError(t, e.Close(), "Close is idempotent")

	b, err := os.ReadFile(out.Name())
	require.NoError(t, err)

	return b
}

func decodeBytes(t *testing.T, b []byte) (*wav.Decoder, []float64) {
	t.Helper()

	d, err := wav.NewDecoder(bytes.NewReader(b))
	require.NoError(t, err)

	got, err := readAllSamples(t, d, 1024*d.Format().Channels)
	require.NoError(t, err)

	return d, got
}

// TestEncoderMatchesBuilder checks the Encoder's output byte for byte against the independent test builder, which
// writes the canonical layout for each fmt variant.
func TestEncoderMatchesBuilder(t *testing.T) {
	ints16 := testutil.IntExtremes(16)
	ints24 := testutil.IntExtremes(24)
	floats := []float64{0, 0.5, -0.5, 1, -1, 0.25}
	fact := func(frames uint32) []testutil.Chunk {
		return []testutil.Chunk{{ID: "fact", Data: binary.LittleEndian.AppendUint32(nil, frames)}}
	}

	for _, tc := range []struct {
		name    string
		format  audio.Format
		samples []float64
		want    []byte
	}{
		{
			name:    "PCM16Stereo",
			format:  audio.Format{SampleRate: 44100, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM},
			samples: scaled(16, testutil.Interleave(ints16, ints16)...),
			want:    testutil.WAV{Channels: 2, SampleRate: 44100, BitsPerSample: 16}.EncodeInts([][]int32{ints16, ints16}),
		},
		{
			name:    "PCM8MonoOddLength",
			format:  audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingPCM},
			samples: scaled(8, testutil.IntExtremes(8)...),
			want:    testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8}.EncodeInts([][]int32{testutil.IntExtremes(8)}),
		},
		{
			name:    "PCM24MonoIsExtensible",
			format:  audio.Format{SampleRate: 48000, Channels: 1, BitDepth: 24, Encoding: audio.EncodingPCM},
			samples: scaled(24, ints24...),
			want: testutil.WAV{Format: testutil.FormatExtensible, Channels: 1, SampleRate: 48000, BitsPerSample: 24}.
				EncodeInts([][]int32{ints24}),
		},
		{
			name: "PCM16FivePointOneIsExtensible",
			format: audio.Format{
				SampleRate: 48000, Channels: 6, BitDepth: 16, Encoding: audio.EncodingPCM, ChannelMask: 0x3F,
			},
			samples: scaled(16, testutil.Interleave(ints16, ints16, ints16, ints16, ints16, ints16)...),
			want: testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 6, SampleRate: 48000, BitsPerSample: 16, ChannelMask: 0x3F,
			}.EncodeInts([][]int32{ints16, ints16, ints16, ints16, ints16, ints16}),
		},
		{
			name: "PCM20In32",
			format: audio.Format{
				SampleRate: 96000, Channels: 2, BitDepth: 32, ValidBits: 20, Encoding: audio.EncodingPCM, ChannelMask: 0x3,
			},
			samples: scaled(32, 1<<12, -1<<12, math.MinInt32, math.MaxInt32-(1<<12-1)),
			want: testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 2, SampleRate: 96000, BitsPerSample: 32, ValidBits: 20,
				ChannelMask: 0x3,
			}.EncodeInts([][]int32{{1 << 12, math.MinInt32}, {-1 << 12, math.MaxInt32 - (1<<12 - 1)}}),
		},
		{
			name:    "Float32WithFact",
			format:  audio.Format{SampleRate: 48000, Channels: 2, BitDepth: 32, Encoding: audio.EncodingFloat},
			samples: floats,
			want: testutil.WAV{Format: testutil.FormatFloat, Channels: 2, SampleRate: 48000, BitsPerSample: 32, Before: fact(3)}.
				EncodeFloats(testutil.Deinterleave(floats, 2)),
		},
		{
			name:    "Float64ThreeChannelsIsExtensible",
			format:  audio.Format{SampleRate: 48000, Channels: 3, BitDepth: 64, Encoding: audio.EncodingFloat},
			samples: floats,
			want: testutil.WAV{
				Format: testutil.FormatExtensible, SubFormat: testutil.FormatFloat, Channels: 3, SampleRate: 48000,
				BitsPerSample: 64, Before: fact(2),
			}.EncodeFloats(testutil.Deinterleave(floats, 3)),
		},
		{
			name:    "MuLawWithFact",
			format:  audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingMuLaw},
			samples: scaled(16, 0, 32124, -32124),
			want: testutil.WAV{Format: testutil.FormatMuLaw, Channels: 1, SampleRate: 8000, BitsPerSample: 8, Before: fact(3)}.
				Encode([]byte{0xFF, 0x80, 0x00}),
		},
		{
			name:    "ALawWithFact",
			format:  audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingALaw},
			samples: scaled(16, 8, -8, 32256, -32256),
			want: testutil.WAV{Format: testutil.FormatALaw, Channels: 1, SampleRate: 8000, BitsPerSample: 8, Before: fact(4)}.
				Encode([]byte{0xD5, 0x55, 0xAA, 0x2A}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := encodeFile(t, tc.format, tc.samples)
			require.Equal(t, tc.want, got)

			d, samples := decodeBytes(t, got)
			require.Equal(t, tc.format, d.Format())
			requireSameSamples(t, tc.samples, samples)
		})
	}
}

func TestEncoderSaturatesByteRate(t *testing.T) {
	f := audio.Format{SampleRate: math.MaxUint32, Channels: 2, BitDepth: 32, Encoding: audio.EncodingPCM}
	b := encodeFile(t, f, []float64{0.5, -0.5})
	require.Equal(t, uint32(math.MaxUint32), binary.LittleEndian.Uint32(b[fmtOff+8:]))

	d, got := decodeBytes(t, b)
	require.Equal(t, f, d.Format())
	requireSameSamples(t, []float64{0.5, -0.5}, got)
}

func TestEncoderQuantises(t *testing.T) {
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}
	_, got := decodeBytes(t, encodeFile(t, f, []float64{1.5, -1.5, math.NaN(), 1.0 / 65536, 3.0 / 65536}))

	requireSameSamples(t, scaled(16, math.MaxInt16, math.MinInt16, 0, 1, 2), got)
}

func TestEncoderStream(t *testing.T) {
	var buf bytes.Buffer

	f := audio.Format{SampleRate: 44100, Channels: 2, BitDepth: 32, Encoding: audio.EncodingFloat}
	e, err := wav.NewEncoder(&buf, f)
	require.NoError(t, err)

	for i := range 10 {
		_, err := e.WriteSamples([]float64{float64(i) / 10, -float64(i) / 10})
		require.NoError(t, err)
	}

	require.NoError(t, e.Close())

	b := buf.Bytes()
	require.Equal(t, uint32(riff.UnknownSize), binary.LittleEndian.Uint32(b[4:8]), "RIFF size unknown without seeking")

	d, got := decodeBytes(t, b)
	_, ok := d.Frames()
	require.False(t, ok, "data size unknown without seeking")
	require.Len(t, got, 20)
	require.Equal(t, []riff.Chunk{
		{ID: riff.ID("fact"), Size: 4, Data: []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	}, d.Chunks())

	ref, err := testutil.DecodeRef(b)
	require.NoError(t, err, "the reference decoder reads live sizes too")
	require.Equal(t, 10, ref.Frames)
}

func TestEncoderWithChunks(t *testing.T) {
	list := riff.Chunk{ID: riff.ID("LIST"), Size: 5, Data: []byte("INFOx")}
	bext := riff.Chunk{ID: riff.ID("bext"), Size: 2, Data: []byte{1, 2}}
	f := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM}

	b := encodeFile(t, f, []float64{0.5, -0.5}, wav.WithChunks(list, bext))

	ref, err := testutil.DecodeRef(b)
	require.NoError(t, err)
	require.Equal(t, []string{"fmt ", "LIST", "bext", "data"}, ref.ChunkIDs)

	d, got := decodeBytes(t, b)
	require.Equal(t, []riff.Chunk{list, bext}, d.Chunks())
	requireSameSamples(t, []float64{0.5, -0.5}, got)

	// transcoding keeps the metadata
	again := encodeFile(t, d.Format(), got, wav.WithChunks(d.Chunks()...))
	require.Equal(t, b, again)

	// a fact chunk is dropped: the Encoder writes its own when the format needs one
	staleFact := riff.Chunk{ID: riff.ID("fact"), Size: 4, Data: []byte{9, 9, 9, 9}}
	require.Equal(t, b, encodeFile(t, f, []float64{0.5, -0.5}, wav.WithChunks(list, staleFact, bext)))

	float := audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 32, Encoding: audio.EncodingFloat}
	fb := encodeFile(t, float, []float64{0.5, -0.5, 0.25}, wav.WithChunks(staleFact, list))

	fd, _ := decodeBytes(t, fb)
	require.Equal(t, []riff.Chunk{
		{ID: riff.ID("fact"), Size: 4, Data: []byte{3, 0, 0, 0}},
		list,
	}, fd.Chunks())

	// so a float file transcodes with all of its chunks
	again = encodeFile(t, fd.Format(), []float64{0.5, -0.5, 0.25}, wav.WithChunks(fd.Chunks()...))
	require.Equal(t, fb, again)

	for _, bad := range []riff.Chunk{
		{ID: riff.ID("fmt "), Data: []byte{}},
		{ID: riff.ID("data")},
		{ID: riff.ID("LIST"), Size: 100}, // listed without its data
	} {
		_, err := wav.NewEncoder(io.Discard, f, wav.WithChunks(bad))
		require.ErrorIs(t, err, wav.ErrReservedChunk, bad.ID.String())
	}

	_, err = wav.NewEncoder(io.Discard, f, wav.WithChunks(riff.Chunk{ID: riff.FourCC{0, 1, 2, 3}}))
	require.ErrorIs(t, err, riff.ErrBadID)
}

func TestEncoderErrors(t *testing.T) {
	pcm16 := audio.Format{SampleRate: 8000, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM}

	format := func(rate, channels, bits int, enc audio.Encoding) audio.Format {
		return audio.Format{SampleRate: rate, Channels: channels, BitDepth: bits, Encoding: enc}
	}

	for _, tc := range []struct {
		name   string
		format audio.Format
		want   error
	}{
		{name: "ZeroFormat", want: audio.ErrInvalidFormat},
		{name: "ZeroChannels", format: format(8000, 0, 16, audio.EncodingPCM), want: audio.ErrInvalidFormat},
		{name: "Float16", format: format(8000, 1, 16, audio.EncodingFloat), want: audio.ErrUnsupportedFormat},
		{name: "PCM40", format: format(8000, 1, 40, audio.EncodingPCM), want: audio.ErrUnsupportedFormat},
		{name: "TooManyChannels", format: format(8000, 70000, 8, audio.EncodingPCM), want: audio.ErrUnsupportedFormat},
		{name: "BlockAlignTooLarge", format: format(8000, 20000, 32, audio.EncodingPCM), want: audio.ErrUnsupportedFormat},
		{name: "SampleRateTooLarge", format: format(math.MaxUint32+1, 1, 8, audio.EncodingPCM), want: audio.ErrUnsupportedFormat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := wav.NewEncoder(io.Discard, tc.format)
			require.ErrorIs(t, err, tc.want)
			require.Nil(t, e)
		})
	}

	t.Run("PartialFrame", func(t *testing.T) {
		var buf bytes.Buffer

		e, err := wav.NewEncoder(&buf, pcm16)
		require.NoError(t, err)

		size := buf.Len()

		n, err := e.WriteSamples([]float64{0.1, 0.2, 0.3})
		require.ErrorIs(t, err, wav.ErrPartialFrame)
		require.Zero(t, n)
		require.Equal(t, size, buf.Len(), "nothing written")
	})

	t.Run("AfterClose", func(t *testing.T) {
		e, err := wav.NewEncoder(io.Discard, pcm16)
		require.NoError(t, err)
		require.NoError(t, e.Close())

		_, err = e.WriteSamples([]float64{0, 0})
		require.ErrorIs(t, err, wav.ErrClosed)
	})

	t.Run("HeaderWriteError", func(t *testing.T) {
		_, err := wav.NewEncoder(&failingWriter{left: 30}, pcm16)
		require.ErrorIs(t, err, errWriteFailed)
	})

	t.Run("WriteErrorIsSticky", func(t *testing.T) {
		w := &failingWriter{left: 44 + 6} // header, then one frame and a half
		e, err := wav.NewEncoder(w, pcm16)
		require.NoError(t, err)

		n, err := e.WriteSamples([]float64{0, 0, 0, 0})
		require.ErrorIs(t, err, errWriteFailed)
		require.Equal(t, 2, n, "one whole frame was written")
		require.Equal(t, int64(1), e.Frames())

		_, err = e.WriteSamples([]float64{0, 0})
		require.ErrorIs(t, err, errWriteFailed)
		require.ErrorIs(t, e.Close(), errWriteFailed, "checking Close is enough to know the file is incomplete")
		require.ErrorIs(t, e.Close(), errWriteFailed)
	})
}

var errWriteFailed = errors.New("write failed")

// failingWriter accepts left bytes, then fails.
type failingWriter struct{ left int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if len(p) <= w.left {
		w.left -= len(p)

		return len(p), nil
	}

	n := w.left
	w.left = 0

	return n, errWriteFailed
}
