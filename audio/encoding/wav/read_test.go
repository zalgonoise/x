package wav_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/osc"
)

// dataPayload returns the raw bytes of a file's data chunk (without its pad byte), using the reference decoder's
// chunk walk; the fixtures and builder files used here end with their data chunk or carry trailing chunks we skip.
func dataPayload(t *testing.T, file []byte) []byte {
	t.Helper()

	for off := 12; off+8 <= len(file); {
		id := string(file[off : off+4])
		size := int(binary.LittleEndian.Uint32(file[off+4 : off+8]))

		if id == data.DataIDString {
			end := off + 8 + size
			if size == 0 || size == math.MaxUint32 || end > len(file) {
				end = len(file)
			}

			return file[off+8 : end]
		}

		off += 8 + size + size%2
	}

	t.Fatal("no data chunk")

	return nil
}

// readers wraps a file in every short-read pattern a network or pipe can produce.
func readers(file []byte) map[string]io.Reader {
	out := map[string]io.Reader{
		"bytes.Reader":   bytes.NewReader(file),
		"OneByteReader":  iotest.OneByteReader(bytes.NewReader(file)),
		"HalfReader":     iotest.HalfReader(bytes.NewReader(file)),
		"DataErrReader":  iotest.DataErrReader(bytes.NewReader(file)),
		"DataErr+Chunks": iotest.DataErrReader(testutil.ChunkReader(bytes.NewReader(file), 1001)),
	}

	for _, sizes := range testutil.ReadSizes {
		out["Chunks"+strconv.Itoa(sizes[0])+"x"+strconv.Itoa(len(sizes))] = testutil.ChunkReader(bytes.NewReader(file), sizes...)
	}

	return out
}

func TestFromShortReads(t *testing.T) {
	for _, fx := range fixtures.All(t, nil) {
		want := dataPayload(t, fx.Data)

		for name, r := range readers(fx.Data) {
			t.Run(fx.Name+"/"+name, func(t *testing.T) {
				w, err := wav.From(r)
				require.NoError(t, err)
				require.Equal(t, want, w.Data.Bytes())
			})
		}
	}
}

func TestStreamShortReadsWholeFrames(t *testing.T) {
	for _, fx := range fixtures.All(t, nil) {
		want := dataPayload(t, fx.Data)

		for name, r := range readers(fx.Data) {
			t.Run(fx.Name+"/"+name, func(t *testing.T) {
				var got []float64

				s := wav.NewStream(func(h *wav.Header, samples []float64) error {
					require.Zero(t, len(samples)%int(h.NumChannels), "processor input holds whole frames")
					got = append(got, samples...)

					return nil
				}, wav.WithSize(4096))

				_, err := s.ReadFrom(r)
				require.NoError(t, err)

				ring, ok := s.Data.(*data.Ring)
				require.True(t, ok)
				require.Equal(t, want, ring.Converter.Bytes(got))
			})
		}
	}
}

func TestDecodeChunkLayouts(t *testing.T) {
	samples := [][]int32{{0, 1000, -1000, 32767, -32768, 7}, {5, -5, 300, -300, 1, -1}}

	for _, testcase := range []struct {
		name   string
		file   testutil.WAV
		chunks []string
	}{
		{name: "Canonical", file: testutil.WAV{}, chunks: []string{"data"}},
		{
			name:   "LISTBeforeData",
			file:   testutil.WAV{Before: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOISFT\x05\x00\x00\x00Lavf\x00")}}},
			chunks: []string{"LIST", "data"},
		},
		{
			name:   "OddLISTBeforeData",
			file:   testutil.WAV{Before: []testutil.Chunk{{ID: "LIST", Data: []byte("odd")}}},
			chunks: []string{"LIST", "data"},
		},
		{
			name:   "UppercaseJUNKAndFact",
			file:   testutil.WAV{Before: []testutil.Chunk{{ID: "JUNK", Data: make([]byte, 28)}, {ID: "fact", Data: []byte{6, 0, 0, 0}}}},
			chunks: []string{"JUNK", "fact", "data"},
		},
		{name: "FmtSize18", file: testutil.WAV{FmtSize: 18}, chunks: []string{"data"}},
		{
			name:   "LISTAfterData",
			file:   testutil.WAV{After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOICMT\x01\x00\x00\x00x\x00")}}},
			chunks: []string{"data", "LIST"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.file.Channels, testcase.file.SampleRate, testcase.file.BitsPerSample = 2, 48000, 16
			file := testcase.file.EncodeInts(samples)
			want := testutil.Interleave(samples...)

			check := func(t *testing.T, w *wav.Wav) {
				t.Helper()

				got := w.Data.Value()
				require.Len(t, got, len(want))

				for i := range want {
					require.Equal(t, int(want[i]), got[i], "sample %d", i)
				}

				ids := make([]string, 0, len(w.Chunks))
				for _, c := range w.Chunks {
					ids = append(ids, string(c.Header().Subchunk2ID[:]))
				}

				// From stops reading after the data chunk; Decode (the io.Writer path) keeps trailing chunks too
				require.Subset(t, testcase.chunks, ids)
				require.Contains(t, ids, "data")
			}

			t.Run("Decode", func(t *testing.T) {
				w, err := wav.Decode(file)
				require.NoError(t, err)
				check(t, w)

				// re-encoding keeps every chunk and produces a valid file
				ref, err := testutil.DecodeRef(w.Bytes())
				require.NoError(t, err)
				require.Equal(t, append([]string{"fmt "}, testcase.chunks...), ref.ChunkIDs)
				require.Equal(t, samples, ref.Ints)
			})

			t.Run("From", func(t *testing.T) {
				w, err := wav.From(iotest.OneByteReader(bytes.NewReader(file)))
				require.NoError(t, err)
				check(t, w)
			})
		})
	}
}

func TestWriteSplitAtEveryOffset(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_24bit_44100hz.wav")
	want := dataPayload(t, file)

	for split := 1; split < 200; split++ {
		w := new(wav.Wav)

		_, err := w.Write(bytes.Clone(file[:split]))
		if err != nil {
			require.ErrorIs(t, err, wav.ErrShortDataBuffer, "split %d: only 'need more data' is acceptable", split)
		}

		_, err = w.Write(bytes.Clone(file[split:]))
		require.NoError(t, err, "split %d", split)
		require.Equal(t, want, w.Data.Bytes(), "split %d", split)
	}
}

func TestWriteSplitOddDataThenChunk(t *testing.T) {
	// an odd-sized 8-bit data chunk followed by a LIST chunk: the pad byte may arrive in a later Write than the data
	samples := [][]int32{{-128, -1, 0, 1, 127}}
	file := testutil.WAV{
		Channels: 1, SampleRate: 44100, BitsPerSample: 8,
		After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOICMT\x02\x00\x00\x00ok")}},
	}.EncodeInts(samples)

	want, err := wav.Decode(file)
	require.NoError(t, err)

	for split := 1; split < len(file); split++ {
		w := new(wav.Wav)

		if _, err := w.Write(bytes.Clone(file[:split])); err != nil {
			require.ErrorIs(t, err, wav.ErrShortDataBuffer, "split %d", split)
		}

		_, err := w.Write(bytes.Clone(file[split:]))
		require.NoError(t, err, "split %d", split)
		require.Equal(t, want.Data.Value(), w.Data.Value(), "split %d", split)
		require.Len(t, w.Chunks, 2, "split %d: data and LIST", split)
		require.Equal(t, "LIST", string(w.Chunks[1].Header().Subchunk2ID[:]), "split %d", split)
		require.Equal(t, file, w.Bytes(), "split %d: byte-exact re-encode", split)
	}
}

func TestEncodeSizesAndPadding(t *testing.T) {
	t.Run("RIFFSizeIsLengthMinus8", func(t *testing.T) {
		w, err := wav.New(44100, 16, 1, wav.PCMFormat)
		require.NoError(t, err)

		w.Generate(osc.SineWave, 440, 10*time.Millisecond)
		b := w.Bytes()
		require.Equal(t, uint32(len(b)-8), binary.LittleEndian.Uint32(b[4:8]))

		ref, err := testutil.DecodeRef(b)
		require.NoError(t, err)
		require.Equal(t, 441, ref.Frames)
	})

	t.Run("BytesFollowGenerate", func(t *testing.T) {
		w, err := wav.New(44100, 16, 1, wav.PCMFormat)
		require.NoError(t, err)

		w.Generate(osc.SineWave, 440, 10*time.Millisecond)
		first := len(w.Bytes())

		w.Generate(osc.SineWave, 440, 10*time.Millisecond)
		require.Equal(t, first+441*2, len(w.Bytes()))
	})

	t.Run("OddDataIsPadded", func(t *testing.T) {
		file := fixtures.WAV(t, "amen_kick/amen_kick_mono_8bit_176400hz.wav")

		w, err := wav.Decode(file)
		require.NoError(t, err)
		require.Len(t, w.Data.Float(), 37393, "the pad byte is not a sample")
		require.Equal(t, file, w.Bytes(), "byte-exact round trip, pad byte included")
	})

	t.Run("TruncatedInputIsNotZeroPadded", func(t *testing.T) {
		file := fixtures.WAV(t, "amen_kick/amen_kick_mono_16bit_44100hz.wav")

		w, err := wav.From(bytes.NewReader(file[:len(file)-1000]))
		require.NoError(t, err)

		ref, err := testutil.DecodeRef(w.Bytes())
		require.NoError(t, err)
		require.Equal(t, uint32(ref.Frames*2), ref.DataSize, "declared size follows the decoded samples")
	})
}

func TestStreamHookSeesEveryByteOnce(t *testing.T) {
	file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_16bit_44100hz.wav")

	var seen []byte

	s := wav.NewStream(func(*wav.Header, []float64) error { return nil }, wav.WithSize(1024),
		wav.WithHook(func(_ context.Context, _ *wav.Header, b []byte) error {
			seen = append(seen, b...)

			return nil
		}))

	_, err := s.ReadFrom(iotest.DataErrReader(testutil.ChunkReader(bytes.NewReader(file), 7, 1000)))
	require.NoError(t, err)
	require.Equal(t, file, seen)
}

func TestHookZeroValues(t *testing.T) {
	n, err := wav.ReaderHook{Reader: bytes.NewReader([]byte("abc"))}.Read(make([]byte, 3))
	require.NoError(t, err)
	require.Equal(t, 3, n)

	_, err = wav.ReaderContextHook{}.Read(make([]byte, 3))
	require.ErrorIs(t, err, io.EOF)
}

func TestBufferSizeHelpers(t *testing.T) {
	const byteRate = 176400 // 44.1 kHz, 16-bit, stereo

	require.Equal(t, byteRate, wav.TimeToBufferSize(byteRate, time.Second))
	require.Equal(t, byteRate/2, wav.TimeToBufferSize(byteRate, 500*time.Millisecond))
	require.Equal(t, 4*3600*byteRate, wav.TimeToBufferSize(byteRate, 4*time.Hour), "long durations do not overflow")
	require.Zero(t, wav.TimeToBufferSize(byteRate, 0))

	require.Equal(t, byteRate, wav.RatioToBufferSize(byteRate, 1.0))
	require.Equal(t, byteRate/2, wav.RatioToBufferSize(byteRate, 0.5))
	require.Equal(t, byteRate, wav.RatioToBufferSize(byteRate, 0), "a non-positive ratio is one second")

	t.Run("StreamSizeIsFrameAligned", func(t *testing.T) {
		file := fixtures.WAV(t, "amen_kick/amen_kick_stereo_24bit_44100hz.wav")

		for name, opt := range map[string]cfg.Option[wav.Config]{
			"WithSize(16)":       wav.WithSize(16),
			"WithSize(1000)":     wav.WithSize(1000),
			"WithRatio(1)":       wav.WithRatio(1),
			"WithDuration(10ms)": wav.WithDuration(10 * time.Millisecond),
		} {
			t.Run(name, func(t *testing.T) {
				s := wav.NewStream(func(*wav.Header, []float64) error { return nil }, opt)

				_, err := s.ReadFrom(bytes.NewReader(file))
				require.NoError(t, err)
				require.Zero(t, s.Size%6, "Size %d is a multiple of BlockAlign", s.Size)
				require.GreaterOrEqual(t, s.Size, 64)
			})
		}
	})
}
