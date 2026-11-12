package wav_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/riff"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// fmtOff is the offset of the fmt chunk body in a builder-made file: "RIFF" size "WAVE" "fmt " size.
const fmtOff = 20

// refSamples decodes b with the reference decoder and returns its interleaved samples, normalised by 2^(n-1).
func refSamples(tb testing.TB, b []byte) []float64 {
	tb.Helper()

	ref, err := testutil.DecodeRef(b)
	require.NoError(tb, err)

	scale := math.Ldexp(1, int(ref.BitsPerSample)-1)
	out := make([]float64, 0, ref.Frames*int(ref.Channels))

	for i := range ref.Frames {
		for c := range int(ref.Channels) {
			if ref.Ints != nil {
				out = append(out, float64(ref.Ints[c][i])/scale)

				continue
			}

			out = append(out, ref.Floats[c][i])
		}
	}

	return out
}

// readAllSamples drains d with a destination of dstLen samples, checking the SampleReader contract on the way.
func readAllSamples(tb testing.TB, d *wav.Decoder, dstLen int) ([]float64, error) {
	tb.Helper()

	var (
		out []float64
		dst = make([]float64, dstLen)
		ch  = d.Format().Channels
	)

	for {
		n, err := d.ReadSamples(dst)
		if n%ch != 0 {
			tb.Fatalf("ReadSamples returned %d samples, not a multiple of %d channels", n, ch)
		}

		out = append(out, dst[:n]...)

		if errors.Is(err, io.EOF) {
			return out, nil
		}

		if err != nil {
			return out, err
		}
	}
}

func requireSameSamples(tb testing.TB, want, got []float64) {
	tb.Helper()

	require.Len(tb, got, len(want))

	for i := range want {
		if math.IsNaN(want[i]) && math.IsNaN(got[i]) {
			continue
		}

		if math.Float64bits(want[i]) != math.Float64bits(got[i]) {
			tb.Fatalf("sample %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func scaled(bits int, vs ...int32) []float64 {
	out := make([]float64, len(vs))
	for i, v := range vs {
		out[i] = float64(v) / math.Ldexp(1, bits-1)
	}

	return out
}

func TestDecoderFixtures(t *testing.T) {
	for _, f := range fixtures.All(t, nil) {
		t.Run(f.Name, func(t *testing.T) {
			d, err := wav.NewDecoder(bytes.NewReader(f.Data))
			require.NoError(t, err)

			format := d.Format()
			require.Equal(t, int(f.SampleRate), format.SampleRate)
			require.Equal(t, int(f.Channels), format.Channels)
			require.Equal(t, int(f.BitsPerSample), format.BitDepth)
			require.Equal(t, audio.EncodingPCM, format.Encoding)

			frames, ok := d.Frames()
			require.True(t, ok)
			require.Equal(t, int64(f.Frames), frames)

			got, err := readAllSamples(t, d, 4096*format.Channels)
			require.NoError(t, err)
			requireSameSamples(t, refSamples(t, f.Data), got)
		})
	}
}

// TestDecoderShortReads decodes every fixture through readers that split headers, chunk boundaries and samples, with
// destinations of one frame, an odd number of frames and a non-multiple of the channel count.
func TestDecoderShortReads(t *testing.T) {
	readers := map[string]func(io.Reader) io.Reader{
		"OneByte": iotest.OneByteReader,
		"Half":    iotest.HalfReader,
		"DataErr": iotest.DataErrReader,
	}

	for _, sizes := range [][]int{{3}, {5}, {7}, {1023}, {4097}, {1, 2, 3, 5, 8, 13}} {
		readers[fmt.Sprint("Chunk", sizes)] = func(r io.Reader) io.Reader { return testutil.ChunkReader(r, sizes...) }
	}

	for _, f := range fixtures.All(t, nil) {
		want := refSamples(t, f.Data)
		ch := int(f.Channels)

		for name, wrap := range readers {
			for _, dstLen := range []int{ch, 7 * ch, 7*ch + 1, 1000 * ch} {
				t.Run(fmt.Sprintf("%s/%s/dst%d", f.Name, name, dstLen), func(t *testing.T) {
					d, err := wav.NewDecoder(wrap(bytes.NewReader(f.Data)))
					require.NoError(t, err)

					got, err := readAllSamples(t, d, dstLen)
					require.NoError(t, err)
					requireSameSamples(t, want, got)
				})
			}
		}
	}
}

func TestDecoderFormats(t *testing.T) {
	ints8 := testutil.IntExtremes(8)
	ints16 := testutil.IntExtremes(16)
	ints24 := testutil.IntExtremes(24)
	ints32 := testutil.IntExtremes(32)
	floats := []float64{0, 0.5, -0.5, 1, -1, 1.5, -2, math.SmallestNonzeroFloat64, math.NaN(), math.Inf(1)}

	shift := func(vs []int32, by int) []int32 {
		out := make([]int32, len(vs))
		for i, v := range vs {
			out[i] = v << by
		}

		return out
	}

	f32 := func(vs []float64) []float64 {
		out := make([]float64, len(vs))
		for i, v := range vs {
			out[i] = float64(float32(v))
		}

		return out
	}

	for _, tc := range []struct {
		name   string
		file   []byte
		format audio.Format
		want   []float64
	}{
		{
			name:   "PCM8Mono",
			file:   testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8}.EncodeInts([][]int32{ints8}),
			format: audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingPCM},
			want:   scaled(8, ints8...),
		},
		{
			name: "PCM16StereoFmt18",
			file: testutil.WAV{Channels: 2, SampleRate: 22050, BitsPerSample: 16, FmtSize: 18}.
				EncodeInts([][]int32{ints16, ints16}),
			format: audio.Format{SampleRate: 22050, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM},
			want:   scaled(16, testutil.Interleave(ints16, ints16)...),
		},
		{
			name: "PCM24SixChannels3000Hz",
			file: testutil.WAV{Channels: 6, SampleRate: 3000, BitsPerSample: 24}.
				EncodeInts([][]int32{ints24, ints24, ints24, ints24, ints24, ints24}),
			format: audio.Format{SampleRate: 3000, Channels: 6, BitDepth: 24, Encoding: audio.EncodingPCM},
			want:   scaled(24, testutil.Interleave(ints24, ints24, ints24, ints24, ints24, ints24)...),
		},
		{
			name:   "PCM32ThreeChannels384kHz",
			file:   testutil.WAV{Channels: 3, SampleRate: 384000, BitsPerSample: 32}.EncodeInts([][]int32{ints32, ints32, ints32}),
			format: audio.Format{SampleRate: 384000, Channels: 3, BitDepth: 32, Encoding: audio.EncodingPCM},
			want:   scaled(32, testutil.Interleave(ints32, ints32, ints32)...),
		},
		{
			name:   "PCM16Fmt40NotExtensible",
			file:   testutil.WAV{Channels: 1, SampleRate: 1, BitsPerSample: 16, FmtSize: 40}.EncodeInts([][]int32{ints16}),
			format: audio.Format{SampleRate: 1, Channels: 1, BitDepth: 16, Encoding: audio.EncodingPCM},
			want:   scaled(16, ints16...),
		},
		{
			name: "Float32",
			file: testutil.WAV{Format: testutil.FormatFloat, Channels: 1, SampleRate: 48000, BitsPerSample: 32}.
				EncodeFloats([][]float64{floats}),
			format: audio.Format{SampleRate: 48000, Channels: 1, BitDepth: 32, Encoding: audio.EncodingFloat},
			want:   f32(floats),
		},
		{
			name: "Float64Stereo",
			file: testutil.WAV{Format: testutil.FormatFloat, Channels: 2, SampleRate: 96000, BitsPerSample: 64}.
				EncodeFloats([][]float64{floats, floats}),
			format: audio.Format{SampleRate: 96000, Channels: 2, BitDepth: 64, Encoding: audio.EncodingFloat},
			want:   testutil.Interleave(floats, floats),
		},
		{
			name: "Extensible20In32FivePointOne",
			file: testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 6, SampleRate: 48000, BitsPerSample: 32,
				ValidBits: 20, ChannelMask: 0x3F,
			}.EncodeInts([][]int32{
				shift(testutil.IntExtremes(20), 12), shift(testutil.IntExtremes(20), 12), shift(testutil.IntExtremes(20), 12),
				shift(testutil.IntExtremes(20), 12), shift(testutil.IntExtremes(20), 12), shift(testutil.IntExtremes(20), 12),
			}),
			format: audio.Format{
				SampleRate: 48000, Channels: 6, BitDepth: 32, ValidBits: 20, Encoding: audio.EncodingPCM, ChannelMask: 0x3F,
			},
			want: scaled(20, testutil.Interleave(
				testutil.IntExtremes(20), testutil.IntExtremes(20), testutil.IntExtremes(20),
				testutil.IntExtremes(20), testutil.IntExtremes(20), testutil.IntExtremes(20),
			)...),
		},
		{
			name: "ExtensiblePCM16Stereo",
			file: testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 2, SampleRate: 44100, BitsPerSample: 16, ChannelMask: 0x3,
			}.EncodeInts([][]int32{ints16, ints16}),
			format: audio.Format{SampleRate: 44100, Channels: 2, BitDepth: 16, Encoding: audio.EncodingPCM, ChannelMask: 0x3},
			want:   scaled(16, testutil.Interleave(ints16, ints16)...),
		},
		{
			name: "ExtensibleFloat32",
			file: testutil.WAV{
				Format: testutil.FormatExtensible, SubFormat: testutil.FormatFloat, Channels: 1, SampleRate: 44100,
				BitsPerSample: 32,
			}.EncodeFloats([][]float64{floats}),
			format: audio.Format{SampleRate: 44100, Channels: 1, BitDepth: 32, Encoding: audio.EncodingFloat},
			want:   f32(floats),
		},
		{
			name: "PCM12In16NotExtensible",
			file: func() []byte {
				b := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{
					shift(testutil.IntExtremes(12), 4),
				})
				binary.LittleEndian.PutUint16(b[fmtOff+14:], 12) // wBitsPerSample: 12 bits in 16-bit containers

				return b
			}(),
			format: audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 16, ValidBits: 12, Encoding: audio.EncodingPCM},
			want:   scaled(12, testutil.IntExtremes(12)...),
		},
		{
			// reference values from the Sun g711.c decoder
			name: "ALaw",
			file: testutil.WAV{Format: testutil.FormatALaw, Channels: 1, SampleRate: 8000, BitsPerSample: 8}.
				Encode([]byte{0xD5, 0x55, 0xAA, 0x2A}),
			format: audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingALaw},
			want:   scaled(16, 8, -8, 32256, -32256),
		},
		{
			name: "MuLaw",
			file: testutil.WAV{Format: testutil.FormatMuLaw, Channels: 1, SampleRate: 8000, BitsPerSample: 8}.
				Encode([]byte{0xFF, 0x7F, 0x80, 0x00}),
			format: audio.Format{SampleRate: 8000, Channels: 1, BitDepth: 8, Encoding: audio.EncodingMuLaw},
			want:   scaled(16, 0, 0, 32124, -32124),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := wav.NewDecoder(testutil.ChunkReader(bytes.NewReader(tc.file), 3))
			require.NoError(t, err)
			require.Equal(t, tc.format, d.Format())

			got, err := readAllSamples(t, d, 5*tc.format.Channels)
			require.NoError(t, err)
			requireSameSamples(t, tc.want, got)
		})
	}
}

func TestDecoderChunks(t *testing.T) {
	samples := [][]int32{{-128, 0, 127}} // 3 bytes of 8-bit data: odd, so a pad byte follows
	before := []testutil.Chunk{{ID: "LIST", Data: []byte("INFOx")}, {ID: "JUNK", Data: []byte{0, 0, 0, 0}}}
	after := []testutil.Chunk{{ID: "id3 ", Data: []byte("tag")}, {ID: "LIST", Data: []byte("INFO")}}
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8, Before: before, After: after}.EncodeInts(samples)

	asChunks := func(cs ...testutil.Chunk) []riff.Chunk {
		out := make([]riff.Chunk, len(cs))
		for i, c := range cs {
			out[i] = riff.Chunk{ID: riff.ID(c.ID), Size: uint32(len(c.Data)), Data: c.Data}
		}

		return out
	}

	t.Run("BeforeAndAfter", func(t *testing.T) {
		d, err := wav.NewDecoder(iotest.OneByteReader(bytes.NewReader(file)))
		require.NoError(t, err)
		require.Equal(t, asChunks(before...), d.Chunks())

		got, err := readAllSamples(t, d, 2)
		require.NoError(t, err)
		requireSameSamples(t, scaled(8, samples[0]...), got)

		require.NoError(t, d.ReadTrailer())
		require.Equal(t, asChunks(append(before, after...)...), d.Chunks())
		require.NoError(t, d.ReadTrailer(), "ReadTrailer is idempotent")
		require.Len(t, d.Chunks(), 4)
	})

	t.Run("TrailerBeforeTheEnd", func(t *testing.T) {
		d, err := wav.NewDecoder(bytes.NewReader(file))
		require.NoError(t, err)

		n, err := d.ReadSamples(make([]float64, 1))
		require.NoError(t, err)
		require.Equal(t, 1, n)

		require.NoError(t, d.ReadTrailer())
		require.Equal(t, asChunks(append(before, after...)...), d.Chunks())

		n, err = d.ReadSamples(make([]float64, 4))
		require.ErrorIs(t, err, io.EOF)
		require.Zero(t, n)
	})

	t.Run("DamagedTrailer", func(t *testing.T) {
		// a chunk header cut short after the last chunk, inside the RIFF body
		damaged := append(bytes.Clone(file), 'L', 'I')
		binary.LittleEndian.PutUint32(damaged[4:], uint32(len(damaged)-8))

		d, err := wav.NewDecoder(bytes.NewReader(damaged))
		require.NoError(t, err)

		got, err := readAllSamples(t, d, 8)
		require.NoError(t, err, "a damaged trailer does not affect the audio")
		requireSameSamples(t, scaled(8, samples[0]...), got)
		require.ErrorIs(t, d.ReadTrailer(), audio.ErrInvalidFormat)
		require.ErrorIs(t, d.ReadTrailer(), audio.ErrInvalidFormat, "the error is sticky")
		require.Equal(t, asChunks(append(before, after...)...), d.Chunks(), "chunks before the damage are kept")

		_, err = d.ReadSamples(make([]float64, 2))
		require.ErrorIs(t, err, audio.ErrInvalidFormat)
	})

	t.Run("BytesAfterTheRIFFBody", func(t *testing.T) {
		for name, tail := range map[string][]byte{
			"ID3v1":     append([]byte("TAG"), make([]byte, 125)...),
			"Padding":   make([]byte, 16),
			"Truncated": []byte("LI"),
			"NextFile":  file,
		} {
			r := bytes.NewReader(append(bytes.Clone(file), tail...))

			d, err := wav.NewDecoder(r)
			require.NoError(t, err)

			_, err = readAllSamples(t, d, 8)
			require.NoError(t, err)
			require.NoError(t, d.ReadTrailer(), name)
			require.Equal(t, asChunks(append(before, after...)...), d.Chunks(), name)
			require.Equal(t, len(tail), r.Len(), "%s: nothing past the RIFF body is read", name)
		}
	})

	for _, tc := range []struct {
		name  string
		limit int
		want  []riff.Chunk
	}{
		{name: "LimitKeepsFirstChunk", limit: wav.ChunkOverhead + 5, want: asChunks(before[0])},
		{
			name:  "LimitListsWithoutData",
			limit: wav.ChunkOverhead + 4,
			want:  []riff.Chunk{{ID: riff.ID("LIST"), Size: 5}},
		},
		{
			name:  "LimitListsSecondWithoutData",
			limit: wav.ChunkOverhead + 5 + wav.ChunkOverhead + 3,
			want:  append(asChunks(before[0]), riff.Chunk{ID: riff.ID("JUNK"), Size: 4}),
		},
		{name: "LimitZero", limit: 0},
		{name: "LimitNegative", limit: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := wav.NewDecoder(bytes.NewReader(file), wav.WithMetadataLimit(tc.limit))
			require.NoError(t, err)
			require.Equal(t, tc.want, d.Chunks())

			got, err := readAllSamples(t, d, 3)
			require.NoError(t, err)
			requireSameSamples(t, scaled(8, samples[0]...), got)
		})
	}
}

func TestDecoderDataBounds(t *testing.T) {
	ints := testutil.IntExtremes(16)
	stereo := testutil.WAV{Channels: 2, SampleRate: 48000, BitsPerSample: 16}
	want := scaled(16, testutil.Interleave(ints, ints)...)

	for _, size := range []uint32{0, math.MaxUint32} {
		t.Run(fmt.Sprintf("LiveSize%#x", size), func(t *testing.T) {
			w := stereo
			w.DataSize, w.RIFFSize = &size, &size

			d, err := wav.NewDecoder(testutil.ChunkReader(bytes.NewReader(w.EncodeInts([][]int32{ints, ints})), 5))
			require.NoError(t, err)

			_, ok := d.Frames()
			require.False(t, ok)

			got, err := readAllSamples(t, d, 6)
			require.NoError(t, err)
			requireSameSamples(t, want, got)
			require.NoError(t, d.ReadTrailer())
		})
	}

	t.Run("TruncatedMidFrame", func(t *testing.T) {
		file := stereo.EncodeInts([][]int32{ints, ints})
		d, err := wav.NewDecoder(bytes.NewReader(file[:len(file)-3]))
		require.NoError(t, err)

		frames, ok := d.Frames()
		require.True(t, ok)
		require.Equal(t, int64(len(ints)), frames, "Frames reports the declared size")

		got, err := readAllSamples(t, d, 4)
		require.NoError(t, err)
		requireSameSamples(t, want[:len(want)-2], got)
	})

	t.Run("PartialFrameBeforePad", func(t *testing.T) {
		// 5 bytes of 16-bit stereo data: one frame and a stray byte, then the pad byte and a LIST chunk
		w := stereo
		w.After = []testutil.Chunk{{ID: "LIST", Data: []byte("INFO")}}
		file := w.Encode([]byte{0x01, 0x00, 0xFF, 0xFF, 0x7F})

		d, err := wav.NewDecoder(bytes.NewReader(file))
		require.NoError(t, err)

		got, err := readAllSamples(t, d, 2)
		require.NoError(t, err)
		requireSameSamples(t, scaled(16, 1, -1), got)
		require.NoError(t, d.ReadTrailer())
		require.Len(t, d.Chunks(), 1)
	})
}

func TestDecoderReadSamples(t *testing.T) {
	ints := testutil.IntExtremes(16)
	file := testutil.WAV{Channels: 2, SampleRate: 48000, BitsPerSample: 16}.EncodeInts([][]int32{ints, ints})

	t.Run("ShortDestination", func(t *testing.T) {
		d, err := wav.NewDecoder(bytes.NewReader(file))
		require.NoError(t, err)

		n, err := d.ReadSamples(nil)
		require.NoError(t, err)
		require.Zero(t, n)

		n, err = d.ReadSamples(make([]float64, 1))
		require.ErrorIs(t, err, io.ErrShortBuffer)
		require.Zero(t, n)

		got, err := readAllSamples(t, d, 2)
		require.NoError(t, err, "a short destination does not consume input")
		require.Len(t, got, 2*len(ints))
	})

	t.Run("EOFIsSticky", func(t *testing.T) {
		d, err := wav.NewDecoder(bytes.NewReader(file))
		require.NoError(t, err)

		_, err = readAllSamples(t, d, 64)
		require.NoError(t, err)

		for range 2 {
			n, err := d.ReadSamples(make([]float64, 2))
			require.ErrorIs(t, err, io.EOF)
			require.Zero(t, n)
		}
	})

	t.Run("ReadErrorIsSticky", func(t *testing.T) {
		boom := errors.New("boom")
		d, err := wav.NewDecoder(io.MultiReader(bytes.NewReader(file[:len(file)-5]), iotest.ErrReader(boom)))
		require.NoError(t, err)

		got, err := readAllSamples(t, d, 2)
		require.ErrorIs(t, err, boom)
		requireSameSamples(t, scaled(16, testutil.Interleave(ints, ints)...)[:len(got)], got)
		require.Len(t, got, 2*len(ints)-4, "every whole frame before the error is returned")

		_, err = d.ReadSamples(make([]float64, 2))
		require.ErrorIs(t, err, boom)
		require.ErrorIs(t, d.ReadTrailer(), boom)
	})
}

func TestDecoderErrors(t *testing.T) {
	pcm16 := testutil.WAV{Channels: 2, SampleRate: 48000, BitsPerSample: 16}
	valid := pcm16.EncodeInts([][]int32{{1, 2}, {3, 4}})

	patch := func(b []byte, off int, v uint16) []byte {
		b = bytes.Clone(b)
		binary.LittleEndian.PutUint16(b[off:], v)

		return b
	}

	fmtBody := valid[fmtOff : fmtOff+16]
	boom := errors.New("boom")

	for _, tc := range []struct {
		name    string
		input   io.Reader
		want    error
		invalid bool // also matches audio.ErrInvalidFormat
		unsupp  bool // also matches audio.ErrUnsupportedFormat
	}{
		{name: "Empty", input: bytes.NewReader(nil), want: wav.ErrNotWAV},
		{name: "ShortRIFF", input: bytes.NewReader(valid[:11]), want: wav.ErrNotWAV},
		{name: "RIFX", input: bytes.NewReader(append([]byte("RIFX"), valid[4:]...)), want: wav.ErrNotWAV},
		{name: "AVI", input: bytes.NewReader(append([]byte("RIFF\x04\x00\x00\x00AVI "), valid[12:]...)), want: wav.ErrNotWAV},
		{name: "NotRIFF", input: bytes.NewReader([]byte("not a wav file at all")), want: wav.ErrNotWAV},
		{
			name:  "DataBeforeFmt",
			input: bytes.NewReader([]byte("RIFF\x00\x00\x00\x00WAVEdata\x02\x00\x00\x00\x01\x02")),
			want:  wav.ErrNoFormat, invalid: true,
		},
		{name: "FmtOnly", input: bytes.NewReader(valid[:fmtOff+16]), want: wav.ErrNoData, invalid: true},
		{name: "FmtTruncatedByEOF", input: bytes.NewReader(valid[:fmtOff+10]), want: audio.ErrInvalidFormat, invalid: true},
		{
			name: "FmtTooShort",
			input: bytes.NewReader(testutil.WAV{Channels: 2, SampleRate: 48000, BitsPerSample: 16, FmtSize: 14}.
				EncodeInts([][]int32{{1}, {2}})),
			want: audio.ErrInvalidFormat, invalid: true,
		},
		{name: "ZeroChannels", input: bytes.NewReader(patch(valid, fmtOff+2, 0)), want: audio.ErrInvalidFormat, invalid: true},
		{name: "ZeroRate", input: bytes.NewReader(patch(patch(valid, fmtOff+4, 0), fmtOff+6, 0)), want: audio.ErrInvalidFormat, invalid: true},
		{name: "ZeroBits", input: bytes.NewReader(patch(valid, fmtOff+14, 0)), want: audio.ErrInvalidFormat, invalid: true},
		{name: "BlockAlignZero", input: bytes.NewReader(patch(valid, fmtOff+12, 0)), want: audio.ErrInvalidFormat, invalid: true},
		{name: "BlockAlignWrong", input: bytes.NewReader(patch(valid, fmtOff+12, 3)), want: audio.ErrInvalidFormat, invalid: true},
		{
			name: "ExtensibleTooShort",
			input: bytes.NewReader(testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 1, SampleRate: 8000,
				BitsPerSample: 16, FmtSize: 18,
			}.EncodeInts([][]int32{{1}})),
			want: audio.ErrInvalidFormat, invalid: true,
		},
		{
			name: "ValidBitsAboveContainer",
			input: bytes.NewReader(testutil.WAV{
				Format: testutil.FormatExtensible, Channels: 1, SampleRate: 8000,
				BitsPerSample: 16, ValidBits: 17,
			}.EncodeInts([][]int32{{1}})),
			want: audio.ErrInvalidFormat, invalid: true,
		},
		{
			name: "SecondFmt",
			input: bytes.NewReader(testutil.WAV{
				Channels: 2, SampleRate: 48000, BitsPerSample: 16,
				Before: []testutil.Chunk{{ID: "fmt ", Data: fmtBody}},
			}.EncodeInts([][]int32{{1}, {2}})),
			want: audio.ErrInvalidFormat, invalid: true,
		},
		{
			name:  "TruncatedChunkHeader",
			input: bytes.NewReader(append(bytes.Clone(valid[:fmtOff+16]), 'L', 'I')),
			want:  riff.ErrTruncated, invalid: true,
		},
		{
			name: "BadChunkID",
			input: bytes.NewReader(testutil.WAV{
				Channels: 2, SampleRate: 48000, BitsPerSample: 16,
				Before: []testutil.Chunk{{ID: "\x00\x01\x02\x03", Data: []byte{1}}},
			}.EncodeInts([][]int32{{1}, {2}})),
			want: riff.ErrBadID, invalid: true,
		},
		{name: "ADPCM", input: bytes.NewReader(patch(valid, fmtOff, 0x0002)), want: audio.ErrUnsupportedFormat, unsupp: true},
		{
			name: "Float16",
			input: bytes.NewReader(testutil.WAV{
				Format: testutil.FormatFloat, Channels: 1, SampleRate: 8000,
				BitsPerSample: 16,
			}.Encode([]byte{0, 0})),
			want: audio.ErrUnsupportedFormat, unsupp: true,
		},
		{
			name: "PCM40",
			input: bytes.NewReader(testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 40}.
				Encode([]byte{0, 0, 0, 0, 0})),
			want: audio.ErrUnsupportedFormat, unsupp: true,
		},
		{
			name: "ALaw16",
			input: bytes.NewReader(testutil.WAV{
				Format: testutil.FormatALaw, Channels: 1, SampleRate: 8000,
				BitsPerSample: 16,
			}.Encode([]byte{0, 0})),
			want: audio.ErrUnsupportedFormat, unsupp: true,
		},
		{
			name: "ExtensibleUnknownGUID",
			input: bytes.NewReader(func() []byte {
				b := testutil.WAV{Format: testutil.FormatExtensible, Channels: 1, SampleRate: 8000, BitsPerSample: 16}.
					EncodeInts([][]int32{{1}})
				b[fmtOff+39] ^= 0xFF // last byte of the SubFormat GUID

				return b
			}()),
			want: audio.ErrUnsupportedFormat, unsupp: true,
		},
		{
			name: "ExtensibleSubFormatADPCM",
			input: bytes.NewReader(testutil.WAV{
				Format: testutil.FormatExtensible, SubFormat: 0x0002, Channels: 1,
				SampleRate: 8000, BitsPerSample: 16,
			}.EncodeInts([][]int32{{1}})),
			want: audio.ErrUnsupportedFormat, unsupp: true,
		},
		{name: "ReadErrorInFmt", input: io.MultiReader(bytes.NewReader(valid[:fmtOff+4]), iotest.ErrReader(boom)), want: boom},
		{name: "ReadErrorInChunkHeader", input: io.MultiReader(bytes.NewReader(valid[:fmtOff+18]), iotest.ErrReader(boom)), want: boom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := wav.NewDecoder(tc.input)
			require.ErrorIs(t, err, tc.want)
			require.Nil(t, d)
			require.Equal(t, tc.invalid, errors.Is(err, audio.ErrInvalidFormat), "matches ErrInvalidFormat: %v", err)
			require.Equal(t, tc.unsupp, errors.Is(err, audio.ErrUnsupportedFormat), "matches ErrUnsupportedFormat: %v", err)
		})
	}

	t.Run("NilReader", func(t *testing.T) {
		_, err := wav.NewDecoder(nil)
		require.ErrorIs(t, err, wav.ErrNotWAV)
	})
}

func TestParseFormatIgnoresByteRate(t *testing.T) {
	b := testutil.WAV{Channels: 2, SampleRate: 48000, BitsPerSample: 16}.EncodeInts([][]int32{{1}, {2}})
	binary.LittleEndian.PutUint32(b[fmtOff+8:], 12345)

	f, err := wav.ParseFormat(b[fmtOff : fmtOff+16])
	require.NoError(t, err)
	require.Equal(t, 192000, f.ByteRate())
}

func TestDecodeRegistered(t *testing.T) {
	f := fixtures.All(t, func(m fixtures.Meta) bool { return m.Channels == 2 })[0]

	sr, name, err := audio.Decode(bytes.NewReader(f.Data))
	require.NoError(t, err)
	require.Equal(t, "wav", name)

	buf, err := audio.ReadAll(sr)
	require.NoError(t, err)
	require.Equal(t, int(f.Channels), buf.Format.Channels)
	requireSameSamples(t, refSamples(t, f.Data), buf.Samples)

	_, _, err = audio.Decode(bytes.NewReader([]byte("RIFF\x04\x00\x00\x00AVI ")))
	require.ErrorIs(t, err, audio.ErrUnknownFormat)

	_, name, err = audio.Decode(bytes.NewReader([]byte("RIFF\x04\x00\x00\x00WAVE")))
	require.Equal(t, "wav", name)
	require.ErrorIs(t, err, wav.ErrNoFormat)
}
