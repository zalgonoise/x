package pcm_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/encoding/pcm"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

func roundTrip(t *testing.T, c pcm.Codec, encoded []byte) {
	t.Helper()

	samples := make([]float64, len(encoded)/c.SampleSize())
	require.Equal(t, len(samples), c.Decode(samples, encoded))
	require.Equal(t, encoded, c.AppendEncode(nil, samples))
}

func TestQuantizeNormalize(t *testing.T) {
	require.Equal(t, int64(32767), pcm.Quantize(1, 16), "+1.0 clips to the largest sample")
	require.Equal(t, int64(-32768), pcm.Quantize(-1, 16), "-1.0 is exactly the smallest sample")
	require.Equal(t, int64(32767), pcm.Quantize(math.Inf(1), 16))
	require.Equal(t, int64(-32768), pcm.Quantize(math.Inf(-1), 16))
	require.Equal(t, int64(0), pcm.Quantize(math.NaN(), 16))
	require.Equal(t, int64(1), pcm.Quantize(0.6/32768, 16), "rounds to nearest")
	require.Equal(t, int64(-1), pcm.Quantize(-0.6/32768, 16))
	require.Equal(t, int64(math.MaxInt32), pcm.Quantize(2, 32))
	require.Equal(t, int64(math.MinInt32), pcm.Quantize(-2, 32))

	require.InDelta(t, -1.0, pcm.Normalize(-32768, 16), 0)
	require.InDelta(t, 32767.0/32768, pcm.Normalize(32767, 16), 0)
	require.InDelta(t, -1.0, pcm.Normalize(math.MinInt32, 32), 0)
}

func TestQuantizeAllWidths(t *testing.T) {
	for bits := 1; bits <= 64; bits++ {
		maxV := int64(math.MaxInt64 >> (64 - bits))
		minV := -maxV - 1

		require.Equal(t, maxV, pcm.Quantize(1, bits), "%d bits", bits)
		require.Equal(t, maxV, pcm.Quantize(math.Inf(1), bits), "%d bits", bits)
		require.Equal(t, minV, pcm.Quantize(-1, bits), "%d bits", bits)
		require.Equal(t, minV, pcm.Quantize(math.Inf(-1), bits), "%d bits", bits)
		require.Zero(t, pcm.Quantize(0, bits), "%d bits", bits)
	}

	require.Equal(t, int64(1)<<62, pcm.Quantize(0.5, 64))
	require.Panics(t, func() { pcm.Quantize(0, 0) })
	require.Panics(t, func() { pcm.Quantize(0, 65) })
}

// bigOrder is a big-endian binary.ByteOrder that is not the binary.BigEndian value (a wrapper).
type bigOrder struct{ binary.ByteOrder }

func TestCustomByteOrder(t *testing.T) {
	order := bigOrder{binary.BigEndian}

	got := make([]float64, 1)
	pcm.Int{Bits: 16, Order: order}.Decode(got, []byte{0x40, 0x00})
	require.InDelta(t, 0.5, got[0], 0, "Int asks the order which byte comes first")
	require.Equal(t, []byte{0x40, 0x00}, pcm.Int{Bits: 16, Order: order}.AppendEncode(nil, []float64{0.5}))
	require.Equal(t, []byte{0x40, 0x00, 0x00}, pcm.Int{Bits: 24, Order: order}.AppendEncode(nil, []float64{0.5}))

	f := pcm.Float{Bits: 32, Order: order}
	enc := f.AppendEncode(nil, []float64{0.5})
	require.Equal(t, math.Float32bits(0.5), binary.BigEndian.Uint32(enc), "encoded in Order")
	require.Equal(t, 1, f.Decode(got, enc))
	require.InDelta(t, 0.5, got[0], 0)

	f64 := pcm.Float{Bits: 64, Order: order}
	require.Equal(t, math.Float64bits(-0.25), binary.BigEndian.Uint64(f64.AppendEncode(nil, []float64{-0.25})))

	native := pcm.Int{Bits: 16, Order: binary.NativeEndian}
	roundTrip(t, native, native.AppendEncode(nil, []float64{0.5, -0.25}))
}

func TestCodecConstructors(t *testing.T) {
	for _, bits := range []int{8, 16, 24, 32} {
		c, err := pcm.NewInt(bits, binary.BigEndian, false)
		require.NoError(t, err)
		require.Equal(t, bits/8, c.SampleSize())
	}

	for _, bits := range []int{0, 4, 12, 40, 64} {
		_, err := pcm.NewInt(bits, nil, false)
		require.ErrorIs(t, err, audio.ErrUnsupportedFormat, "%d bits", bits)
	}

	for _, bits := range []int{32, 64} {
		c, err := pcm.NewFloat(bits, nil)
		require.NoError(t, err)
		require.Equal(t, bits/8, c.SampleSize())
	}

	for _, bits := range []int{0, 16, 24, 128} {
		_, err := pcm.NewFloat(bits, nil)
		require.ErrorIs(t, err, audio.ErrUnsupportedFormat, "%d bits", bits)
	}

	// a misconfigured literal is a programming error: it panics with a clear message instead of misbehaving
	for name, c := range map[string]pcm.Codec{
		"IntZero": pcm.Int{}, "Int64": pcm.Int{Bits: 64}, "FloatZero": pcm.Float{}, "Float16": pcm.Float{Bits: 16},
	} {
		require.Panics(t, func() { c.Decode(make([]float64, 4), make([]byte, 16)) }, name)
		require.Panics(t, func() { c.AppendEncode(nil, []float64{0.5}) }, name)
	}
}

func TestIntRoundTripExhaustive(t *testing.T) {
	u8 := make([]byte, 256)
	for i := range u8 {
		u8[i] = byte(i)
	}

	roundTrip(t, pcm.Int{Bits: 8, Unsigned: true}, u8)
	roundTrip(t, pcm.Int{Bits: 8}, u8)

	s16 := make([]byte, 0, 1<<17)
	for v := math.MinInt16; v <= math.MaxInt16; v++ {
		s16 = binary.LittleEndian.AppendUint16(s16, uint16(int16(v)))
	}

	roundTrip(t, pcm.Int{Bits: 16}, s16)
	roundTrip(t, pcm.Int{Bits: 16, Order: binary.BigEndian}, s16)
}

func TestIntExtremesAgainstBuilder(t *testing.T) {
	for _, bits := range []int{8, 16, 24, 32} {
		want := testutil.IntExtremes(bits)
		file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: uint16(bits)}.EncodeInts([][]int32{want})
		payload := file[44 : 44+len(want)*bits/8]

		codec, err := pcm.WAV(audio.Format{SampleRate: 8000, Channels: 1, BitDepth: bits, Encoding: audio.EncodingPCM})
		require.NoError(t, err)

		got := make([]float64, len(want))
		require.Equal(t, len(want), codec.Decode(got, payload))

		for i := range want {
			require.InDelta(t, pcm.Normalize(int64(want[i]), bits), got[i], 0, "%d-bit sample %d", bits, i)
		}

		require.InDelta(t, -1.0, got[0], 0, "%d-bit minimum is exactly -1", bits)
		require.Equal(t, payload, codec.AppendEncode(nil, got), "%d-bit round trip", bits)
	}
}

func TestIntLayouts(t *testing.T) {
	be24 := pcm.Int{Bits: 24, Order: binary.BigEndian}
	got := make([]float64, 2)
	be24.Decode(got, []byte{0x80, 0x00, 0x00, 0x7F, 0xFF, 0xFF})
	require.Equal(t, []float64{-1, pcm.Normalize(1<<23-1, 24)}, got, "big-endian 24-bit sign extension")

	u8 := pcm.Int{Bits: 8, Unsigned: true}
	require.Equal(t, []byte{0x80, 0xFF, 0x00}, u8.AppendEncode(nil, []float64{0, 1, -1}), "8-bit WAV: 0x80 is silence")

	require.Equal(t, 1, pcm.Int{Bits: 16}.Decode(make([]float64, 4), []byte{1, 2, 3}), "partial sample ignored")
	require.Equal(t, 1, pcm.Int{Bits: 16}.Decode(make([]float64, 1), []byte{1, 2, 3, 4}), "bounded by dst")
}

func TestFloat(t *testing.T) {
	in := []float64{0, 0.5, -0.25, 1, -1, 1.5, -3}

	for _, c := range []pcm.Float{{Bits: 32}, {Bits: 64}, {Bits: 32, Order: binary.BigEndian}, {Bits: 64, Order: binary.BigEndian}} {
		encoded := c.AppendEncode(nil, in)
		require.Len(t, encoded, len(in)*c.SampleSize())

		got := make([]float64, len(in))
		require.Equal(t, len(in), c.Decode(got, encoded))
		require.Equal(t, in, got, "%+v: values outside [-1, 1] are kept", c)
	}

	require.Equal(t, math.Float32bits(0.5), binary.LittleEndian.Uint32(pcm.Float{Bits: 32}.AppendEncode(nil, []float64{0.5})))
}

func TestG711Reference(t *testing.T) {
	decode := func(c pcm.Codec, code byte) int64 {
		var out [1]float64

		c.Decode(out[:], []byte{code})

		return int64(math.Ldexp(out[0], 15))
	}

	// values from the Sun Microsystems g711.c reference implementation
	require.Equal(t, int64(8), decode(pcm.ALaw{}, 0xD5))
	require.Equal(t, int64(-8), decode(pcm.ALaw{}, 0x55))
	require.Equal(t, int64(32256), decode(pcm.ALaw{}, 0xAA))
	require.Equal(t, int64(-32256), decode(pcm.ALaw{}, 0x2A))
	require.Equal(t, int64(0), decode(pcm.MuLaw{}, 0xFF))
	require.Equal(t, int64(0), decode(pcm.MuLaw{}, 0x7F))
	require.Equal(t, int64(32124), decode(pcm.MuLaw{}, 0x80))
	require.Equal(t, int64(-32124), decode(pcm.MuLaw{}, 0x00))

	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}

	samples := make([]float64, 256)

	pcm.ALaw{}.Decode(samples, all)
	require.Equal(t, all, pcm.ALaw{}.AppendEncode(nil, samples), "every A-law code round-trips")

	pcm.MuLaw{}.Decode(samples, all)
	got := pcm.MuLaw{}.AppendEncode(nil, samples)

	for i := range all {
		want := all[i]
		if want == 0x7F {
			want = 0xFF // µ-law has two zero codes; zero encodes as 0xFF
		}

		require.Equal(t, want, got[i], "µ-law code %#02x", i)
	}

	require.Equal(t, []byte{0x2A, 0xAA}, pcm.ALaw{}.AppendEncode(nil, []float64{-1, 1}), "full scale")
	require.Equal(t, []byte{0x00, 0x80}, pcm.MuLaw{}.AppendEncode(nil, []float64{-1, 1}), "full scale clips to the largest code")
}

func TestWAVCodecSelection(t *testing.T) {
	base := audio.Format{SampleRate: 8000, Channels: 1}

	for _, testcase := range []struct {
		enc  audio.Encoding
		bits int
		size int
		err  error
	}{
		{enc: audio.EncodingPCM, bits: 8, size: 1},
		{enc: audio.EncodingPCM, bits: 16, size: 2},
		{enc: audio.EncodingPCM, bits: 24, size: 3},
		{enc: audio.EncodingPCM, bits: 32, size: 4},
		{enc: audio.EncodingFloat, bits: 32, size: 4},
		{enc: audio.EncodingFloat, bits: 64, size: 8},
		{enc: audio.EncodingALaw, bits: 8, size: 1},
		{enc: audio.EncodingMuLaw, bits: 8, size: 1},
		{enc: audio.EncodingPCM, bits: 40, err: audio.ErrUnsupportedFormat},
		{enc: audio.EncodingFloat, bits: 16, err: audio.ErrUnsupportedFormat},
		{enc: audio.EncodingALaw, bits: 16, err: audio.ErrUnsupportedFormat},
		{enc: audio.EncodingPCM, bits: 12, err: audio.ErrInvalidFormat},
	} {
		f := base
		f.Encoding, f.BitDepth = testcase.enc, testcase.bits

		c, err := pcm.WAV(f)
		if testcase.err != nil {
			require.ErrorIs(t, err, testcase.err, "%s", f)

			continue
		}

		require.NoError(t, err, "%s", f)
		require.Equal(t, testcase.size, c.SampleSize(), "%s", f)
	}
}
