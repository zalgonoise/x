package testutil_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net/http"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/internal/testutil"
)

func TestULPDiff(t *testing.T) {
	for _, testcase := range []struct {
		name string
		a, b float64
		want uint64
	}{
		{name: "Equal", a: 1.5, b: 1.5, want: 0},
		{name: "SignedZeroes", a: 0, b: math.Copysign(0, -1), want: 0},
		{name: "Adjacent", a: 1, b: math.Nextafter(1, 2), want: 1},
		{name: "AcrossZero", a: math.Nextafter(0, 1), b: math.Nextafter(0, -1), want: 2},
		{name: "NaN", a: math.NaN(), b: 1, want: math.MaxUint64},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.want, testutil.ULPDiff(testcase.a, testcase.b))
			require.Equal(t, testcase.want, testutil.ULPDiff(testcase.b, testcase.a))
		})
	}
}

func TestNaiveDFT(t *testing.T) {
	const n = 16

	t.Run("Impulse", func(t *testing.T) {
		mags := testutil.Magnitudes(testutil.NaiveDFT(testutil.Real(testutil.Impulse(n, 3))))
		testutil.RequireInDelta(t, testutil.DC(n, 1), mags, 1e-12)
	})

	t.Run("DC", func(t *testing.T) {
		mags := testutil.Magnitudes(testutil.NaiveDFT(testutil.Real(testutil.DC(n, 0.5))))
		want := make([]float64, n)
		want[0] = n * 0.5
		testutil.RequireInDelta(t, want, mags, 1e-12)
	})

	t.Run("BinCentredCosine", func(t *testing.T) {
		mags := testutil.Magnitudes(testutil.NaiveDFT(testutil.Real(testutil.BinCosine(n, 3, 2))))
		want := make([]float64, n)
		want[3], want[n-3] = n, n // amp * n / 2
		testutil.RequireInDelta(t, want, mags, 1e-9)
	})

	t.Run("Alternating", func(t *testing.T) {
		mags := testutil.Magnitudes(testutil.NaiveDFT(testutil.Real(testutil.Alternating(n))))
		want := make([]float64, n)
		want[n/2] = n
		testutil.RequireInDelta(t, want, mags, 1e-9)
	})
}

func TestSignals(t *testing.T) {
	t.Run("SineQuarterPeriod", func(t *testing.T) {
		s := testutil.Sine(5, 4, 1, 1, 0) // 1 Hz at 4 Hz sampling: 0, 1, 0, -1, 0
		testutil.RequireInDelta(t, []float64{0, 1, 0, -1, 0}, s, 1e-12)
	})

	t.Run("NoiseDeterministic", func(t *testing.T) {
		require.Equal(t, testutil.Noise(64, 1, 7), testutil.Noise(64, 1, 7))
		require.NotEqual(t, testutil.Noise(64, 1, 7), testutil.Noise(64, 1, 8))

		for _, v := range testutil.Noise(1024, 0.5, 1) {
			require.True(t, v >= -0.5 && v < 0.5, v)
		}
	})

	t.Run("InterleaveRoundTrip", func(t *testing.T) {
		l, r := []int{1, 2, 3}, []int{-1, -2, -3}
		mixed := testutil.Interleave(l, r)
		require.Equal(t, []int{1, -1, 2, -2, 3, -3}, mixed)
		require.Equal(t, [][]int{l, r}, testutil.Deinterleave(mixed, 2))
	})

	t.Run("IntExtremes", func(t *testing.T) {
		require.Equal(t, []int32{-128, -127, -1, 0, 1, 126, 127}, testutil.IntExtremes(8))
		require.Equal(t, []int32{math.MinInt32, math.MinInt32 + 1, -1, 0, 1, math.MaxInt32 - 1, math.MaxInt32},
			testutil.IntExtremes(32))
	})
}

func TestWAVRoundTrip(t *testing.T) {
	for _, bits := range []uint16{8, 16, 24, 32} {
		for _, channels := range []uint16{1, 2, 6} {
			perChannel := make([][]int32, channels)
			for c := range perChannel {
				// distinct content per channel, so interleaving bugs show up
				perChannel[c] = append(testutil.IntExtremes(int(bits)), int32(c), -int32(c))
			}

			w := testutil.WAV{Channels: channels, SampleRate: 48000, BitsPerSample: bits}
			ref, err := testutil.DecodeRef(w.EncodeInts(perChannel))
			require.NoError(t, err)

			require.Equal(t, testutil.FormatPCM, ref.Format)
			require.Equal(t, channels, ref.Channels)
			require.Equal(t, bits, ref.BitsPerSample)
			require.Equal(t, uint32(48000)*uint32(channels*bits/8), ref.ByteRate)
			require.Equal(t, perChannel, ref.Ints)
			require.Equal(t, []string{"fmt ", "data"}, ref.ChunkIDs)
		}
	}
}

func TestWAVEightBitIsUnsigned(t *testing.T) {
	b := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8}.EncodeInts([][]int32{{-128, 0, 127}})
	require.Equal(t, []byte{0x00, 0x80, 0xFF}, b[len(b)-4:len(b)-1], "8-bit WAV stores value+128")
	require.Zero(t, b[len(b)-1], "odd data chunk is followed by a pad byte")
}

func TestWAVFloat(t *testing.T) {
	for _, bits := range []uint16{32, 64} {
		in := [][]float64{{-1, -0.5, 0, 0.25, 1, 1.5}, {0.125, -0.125, 0, 0, 0, -2}}
		w := testutil.WAV{Format: testutil.FormatFloat, Channels: 2, SampleRate: 44100, BitsPerSample: bits}

		ref, err := testutil.DecodeRef(w.EncodeFloats(in))
		require.NoError(t, err)
		require.Equal(t, testutil.FormatFloat, ref.Format)
		require.Equal(t, in, ref.Floats)
	}
}

func TestWAVChunkLayout(t *testing.T) {
	w := testutil.WAV{
		Channels: 1, SampleRate: 22050, BitsPerSample: 8,
		Before: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOISFT\x05\x00\x00\x00odd!\x00")}, {ID: "JUNK", Data: make([]byte, 4)}},
		After:  []testutil.Chunk{{ID: "id3 ", Data: []byte{1, 2, 3}}},
	}

	samples := [][]int32{{1, 2, 3, 4, 5}} // odd data size: needs a pad byte before "id3 "
	b := w.EncodeInts(samples)

	ref, err := testutil.DecodeRef(b)
	require.NoError(t, err)
	require.Equal(t, []string{"fmt ", "LIST", "JUNK", "data", "id3 "}, ref.ChunkIDs)
	require.Equal(t, samples, ref.Ints)
	require.Equal(t, 5, ref.Frames)
	require.Equal(t, uint32(len(b)-8), uint32(b[4])|uint32(b[5])<<8|uint32(b[6])<<16|uint32(b[7])<<24, "RIFF size = len-8")
}

func TestWAVExtensible(t *testing.T) {
	w := testutil.WAV{
		Format: testutil.FormatExtensible, Channels: 2, SampleRate: 96000, BitsPerSample: 32, ValidBits: 24, ChannelMask: 0x3,
	}

	ref, err := testutil.DecodeRef(w.EncodeInts([][]int32{{1 << 8, -1 << 8}, {2 << 8, -2 << 8}}))
	require.NoError(t, err)
	require.True(t, ref.Extensible)
	require.Equal(t, testutil.FormatPCM, ref.Format)
	require.Equal(t, uint16(24), ref.ValidBits)
	require.Equal(t, uint32(0x3), ref.ChannelMask)
	require.Equal(t, [][]int32{{1 << 8, -1 << 8}, {2 << 8, -2 << 8}}, ref.Ints)
}

func TestWAVLiveSizes(t *testing.T) {
	for _, size := range []uint32{0, math.MaxUint32} {
		w := testutil.WAV{Channels: 1, SampleRate: 44100, BitsPerSample: 16, DataSize: &size, RIFFSize: &size}
		ref, err := testutil.DecodeRef(w.EncodeInts([][]int32{{1, 2, 3}}))
		require.NoError(t, err)
		require.Equal(t, [][]int32{{1, 2, 3}}, ref.Ints, "size %#x decodes until the end of input", size)
	}
}

func TestDecodeRefErrors(t *testing.T) {
	_, err := testutil.DecodeRef([]byte("not a wav file at all"))
	require.ErrorIs(t, err, testutil.ErrRefNotRIFF)

	dataFirst := []byte("RIFF\x0c\x00\x00\x00WAVEdata\x00\x00\x00\x00")
	_, err = testutil.DecodeRef(dataFirst)
	require.ErrorIs(t, err, testutil.ErrRefNoFmt)
}

func TestChunkReader(t *testing.T) {
	src := bytes.Repeat([]byte("0123456789"), 100)

	for _, sizes := range testutil.ReadSizes {
		r := testutil.ChunkReader(bytes.NewReader(src), sizes...)
		buf := make([]byte, 4096)

		for i := 0; ; i++ {
			n, err := r.Read(buf)
			require.LessOrEqual(t, n, max(sizes[i%len(sizes)], 1))

			if errors.Is(err, io.EOF) {
				break
			}

			require.NoError(t, err)
		}

		got, err := io.ReadAll(testutil.ChunkReader(bytes.NewReader(src), sizes...))
		require.NoError(t, err)
		require.Equal(t, src, got)
	}

	require.NoError(t, iotest.TestReader(testutil.ChunkReader(bytes.NewReader(src), 7), src))
}

func TestWAVServer(t *testing.T) {
	body := testutil.WAV{Channels: 1, SampleRate: 44100, BitsPerSample: 16}.EncodeInts([][]int32{testutil.IntExtremes(16)})

	type served struct {
		status int
		header http.Header
		length int64
		body   []byte
		err    error
	}

	get := func(t *testing.T, srv *testutil.WAVServer) served {
		t.Helper()

		res, err := http.Get(srv.URL)
		require.NoError(t, err)

		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)

		return served{status: res.StatusCode, header: res.Header, length: res.ContentLength, body: b, err: err}
	}

	t.Run("Default", func(t *testing.T) {
		srv := testutil.NewWAVServer(t, body)
		res := get(t, srv)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status)
		require.Equal(t, "audio/wav", res.header.Get("Content-Type"))
		require.Equal(t, int64(len(body)), res.length)
		require.Equal(t, body, res.body)
		require.Equal(t, int64(1), srv.Requests())
	})

	t.Run("Chunked", func(t *testing.T) {
		res := get(t, testutil.NewWAVServer(t, body, testutil.Chunked(3)))
		require.NoError(t, res.err)
		require.Equal(t, int64(-1), res.length)
		require.Equal(t, body, res.body)
	})

	t.Run("StatusAndContentType", func(t *testing.T) {
		res := get(t, testutil.NewWAVServer(t, []byte("busy"), testutil.Status(http.StatusServiceUnavailable),
			testutil.ContentType("text/html")))
		require.NoError(t, res.err)
		require.Equal(t, http.StatusServiceUnavailable, res.status)
		require.Equal(t, "text/html", res.header.Get("Content-Type"))
	})

	t.Run("DisconnectAfter", func(t *testing.T) {
		res := get(t, testutil.NewWAVServer(t, body, testutil.DisconnectAfter(20)))
		require.ErrorIs(t, res.err, io.ErrUnexpectedEOF)
		require.Equal(t, body[:20], res.body)
	})

	t.Run("Endless", func(t *testing.T) {
		srv := testutil.NewWAVServer(t, body, testutil.Endless(math.MaxUint32), testutil.Chunked(64))

		res, err := http.Get(srv.URL)
		require.NoError(t, err)

		defer res.Body.Close()

		got := make([]byte, 10*len(body))
		_, err = io.ReadFull(res.Body, got)
		require.NoError(t, err)

		ref, err := testutil.DecodeRef(got)
		require.NoError(t, err)
		require.Equal(t, uint32(math.MaxUint32), ref.DataSize)
		require.Greater(t, ref.Frames, 7*len(testutil.IntExtremes(16)))
		require.Equal(t, testutil.IntExtremes(16), ref.Ints[0][:7], "data payload repeats from the start")
	})
}

// TestWAVServerEndlessOddPayload checks that an endless stream repeats only the audio payload: not the pad byte of an
// odd-sized data chunk, nor the chunks after it.
func TestWAVServerEndlessOddPayload(t *testing.T) {
	samples := []int32{-128, -1, 0, 1, 127} // 8-bit mono: 5 bytes, odd
	file := testutil.WAV{
		Channels: 1, SampleRate: 8000, BitsPerSample: 8,
		After: []testutil.Chunk{{ID: "LIST", Data: []byte("INFOxxxx")}},
	}.EncodeInts([][]int32{samples})

	srv := testutil.NewWAVServer(t, file, testutil.Endless(math.MaxUint32))

	res, err := http.Get(srv.URL)
	require.NoError(t, err)

	defer res.Body.Close()

	got := make([]byte, 44+5*20)
	_, err = io.ReadFull(res.Body, got)
	require.NoError(t, err)

	ref, err := testutil.DecodeRef(got)
	require.NoError(t, err)
	require.Equal(t, uint32(math.MaxUint32), ref.DataSize)

	for i, v := range ref.Ints[0] {
		require.Equal(t, samples[i%len(samples)], v, "sample %d", i)
	}
}

func TestLiveHeader(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16}.EncodeInts([][]int32{{1, 2}})

	live := testutil.LiveHeader(file, 0)
	ref, err := testutil.DecodeRef(live)
	require.NoError(t, err)
	require.Zero(t, ref.DataSize)
	require.Zero(t, binary.LittleEndian.Uint32(live[4:8]), "the RIFF size is a live marker too")
	require.Equal(t, [][]int32{{1, 2}}, ref.Ints, "a 0-size live data chunk runs to the end")
	require.NotEqual(t, file, live)

	// a definite RIFF size that declares chunks after a 0-size data chunk makes it empty
	sized := bytes.Clone(live)
	binary.LittleEndian.PutUint32(sized[4:8], uint32(len(sized)-8))
	ref, err = testutil.DecodeRef(sized[:len(sized)-4])
	require.NoError(t, err)
	require.Empty(t, ref.Ints[0])

	noData := []byte("RIFF\x0c\x00\x00\x00WAVEfmt \x10\x00\x00\x00")
	require.Equal(t, noData, testutil.LiveHeader(noData, 0), "input without a data chunk header is left alone")
}

func TestWAVTruncatedFmt(t *testing.T) {
	file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 16, FmtSize: 14}.EncodeInts([][]int32{{1}})

	_, err := testutil.DecodeRef(file)
	require.ErrorIs(t, err, testutil.ErrRefFormat, "a 14-byte fmt chunk is malformed")
}
