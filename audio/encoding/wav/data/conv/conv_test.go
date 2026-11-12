package conv_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav/data/conv"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

func TestPCM8BitIsUnsigned(t *testing.T) {
	c := conv.PCM8Bit{}

	// normalised by 2^(n-1): the smallest sample is exactly -1, the largest 127/128
	require.Equal(t, []float64{0, 127.0 / 128, -1}, c.Parse([]byte{0x80, 0xFF, 0x00}), "0x80 is silence")
	require.Equal(t, []byte{0x80, 0xFF, 0x00}, c.Bytes([]float64{0, 127.0 / 128, -1}))
	require.Equal(t, []byte{0xFF}, c.Bytes([]float64{1}), "+1.0 clips to the largest sample")
	require.Equal(t, []int{0, 127, -128}, c.Value([]float64{0, 127.0 / 128, -1}))

	t.Run("RoundTripAllValues", func(t *testing.T) {
		all := make([]byte, 256)
		for i := range all {
			all[i] = byte(i)
		}

		require.Equal(t, all, c.Bytes(c.Parse(all)))
	})

	t.Run("MatchesReferenceBuilder", func(t *testing.T) {
		want := testutil.IntExtremes(8)
		file := testutil.WAV{Channels: 1, SampleRate: 8000, BitsPerSample: 8}.EncodeInts([][]int32{want})
		ref, err := testutil.DecodeRef(file)
		require.NoError(t, err)

		got := c.Value(c.Parse(file[44 : 44+len(want)]))
		for i := range want {
			require.Equal(t, int(ref.Ints[0][i]), got[i])
		}
	})
}

func TestPCM16BitRoundTripAllValues(t *testing.T) {
	c := conv.PCM16Bit{}
	all := make([]byte, 0, 1<<17)

	for v := math.MinInt16; v <= math.MaxInt16; v++ {
		all = binary.LittleEndian.AppendUint16(all, uint16(int16(v)))
	}

	require.Equal(t, all, c.Bytes(c.Parse(all)))

	values := c.Value(c.Parse(all))
	for i, v := range values {
		require.Equal(t, math.MinInt16+i, v)
	}
}

func TestPCM24And32BitExtremes(t *testing.T) {
	for _, testcase := range []struct {
		bits int
		c    interface {
			Parse([]byte) []float64
			Bytes([]float64) []byte
			Value([]float64) []int
		}
	}{
		{bits: 24, c: conv.PCM24Bit{}},
		{bits: 32, c: conv.PCM32Bit{}},
	} {
		want := testutil.IntExtremes(testcase.bits)
		file := testutil.WAV{Channels: 1, SampleRate: 48000, BitsPerSample: uint16(testcase.bits)}.EncodeInts([][]int32{want})
		payload := file[44 : 44+len(want)*testcase.bits/8] // without the pad byte of an odd-sized chunk

		got := testcase.c.Value(testcase.c.Parse(payload))
		for i := range want {
			require.Equal(t, int(want[i]), got[i], "%d-bit sample %d", testcase.bits, i)
		}

		require.Equal(t, payload, testcase.c.Bytes(testcase.c.Parse(payload)), "%d-bit round trip", testcase.bits)
	}
}

func TestQuantizeClipsAndRounds(t *testing.T) {
	c16 := conv.PCM16Bit{}

	t.Run("Clip", func(t *testing.T) {
		require.Equal(t, []int{32767, -32768, 32767, -32768}, c16.Value([]float64{1.5, -1.5, math.Inf(1), math.Inf(-1)}))

		b := c16.Bytes([]float64{2, -2})
		require.Equal(t, []byte{0xFF, 0x7F, 0x00, 0x80}, b, "clips to 0x7FFF and 0x8000 instead of wrapping")

		require.Equal(t, []byte{0xFF, 0xFF, 0x7F, 0x00, 0x00, 0x80}, conv.PCM24Bit{}.Bytes([]float64{10, -10}),
			"24-bit clips to 0x7FFFFF and -0x800000")
	})

	t.Run("Round", func(t *testing.T) {
		require.Equal(t, []int{1, -1, 0}, c16.Value([]float64{0.6 / 32767, -0.6 / 32767, 0.4 / 32767}))
	})

	t.Run("NaNIsSilence", func(t *testing.T) {
		require.Equal(t, []int{0}, c16.Value([]float64{math.NaN()}))
		require.Equal(t, []byte{0x80}, conv.PCM8Bit{}.Bytes([]float64{math.NaN()}))
	})

	t.Run("8BitClip", func(t *testing.T) {
		require.Equal(t, []byte{0xFF, 0x00}, conv.PCM8Bit{}.Bytes([]float64{3, -3}))
	})
}

func TestFloatConverters(t *testing.T) {
	in := []float64{0, 0.5, -0.25, 1, -1, 1.5}

	f32 := conv.Float32{}
	require.Equal(t, in, f32.Parse(f32.Bytes(in)))
	require.Equal(t, math.Float32bits(0.5), binary.LittleEndian.Uint32(f32.Bytes([]float64{0.5})))

	f64 := conv.Float64{}
	require.Equal(t, in, f64.Parse(f64.Bytes(in)))
	require.Equal(t, math.Float64bits(-0.25), binary.LittleEndian.Uint64(f64.Bytes([]float64{-0.25})))

	require.Equal(t, []int{math.MaxInt32, math.MinInt32}, f64.Value([]float64{2, -2}), "Value clips instead of overflowing")
}

func TestTrailingPartialSampleIgnored(t *testing.T) {
	require.Len(t, conv.PCM16Bit{}.Parse([]byte{1, 2, 3}), 1)
	require.Len(t, conv.PCM24Bit{}.Parse([]byte{1, 2, 3, 4, 5}), 1)
	require.Len(t, conv.PCM32Bit{}.Parse([]byte{1, 2, 3, 4, 5, 6, 7}), 1)
	require.Len(t, conv.Float32{}.Parse([]byte{1, 2, 3}), 0)
	require.Len(t, conv.Float64{}.Parse([]byte{1, 2, 3, 4, 5, 6, 7}), 0)
}
