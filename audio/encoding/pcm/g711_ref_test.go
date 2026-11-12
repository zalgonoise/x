package pcm_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/pcm"
)

// An independent port of the Sun g711.c reference (13-bit A-law and 14-bit µ-law versions), used only as a test oracle.

func refSearch(val int, table []int) int {
	for i, v := range table {
		if val <= v {
			return i
		}
	}

	return len(table)
}

func refLinearToALaw(v int) byte {
	segEnd := []int{0x1F, 0x3F, 0x7F, 0xFF, 0x1FF, 0x3FF, 0x7FF, 0xFFF}
	v >>= 3

	mask := 0xD5
	if v < 0 {
		mask = 0x55
		v = -v - 1
	}

	seg := refSearch(v, segEnd)
	if seg >= 8 {
		return byte(0x7F ^ mask)
	}

	a := seg << 4
	if seg < 2 {
		a |= (v >> 1) & 0xF
	} else {
		a |= (v >> seg) & 0xF
	}

	return byte(a ^ mask)
}

func refALawToLinear(a byte) int {
	a ^= 0x55
	t := int(a&0x0F) << 4
	seg := int(a&0x70) >> 4

	switch seg {
	case 0:
		t += 8
	case 1:
		t += 0x108
	default:
		t += 0x108
		t <<= seg - 1
	}

	if a&0x80 != 0 {
		return t
	}

	return -t
}

func refLinearToMuLaw(v int) byte {
	const bias = 0x84

	segEnd := []int{0xFF, 0x1FF, 0x3FF, 0x7FF, 0xFFF, 0x1FFF, 0x3FFF, 0x7FFF}

	mask := 0xFF

	if v < 0 {
		v = bias - v
		mask = 0x7F
	} else {
		v += bias
	}

	seg := refSearch(v, segEnd)
	if seg >= 8 {
		return byte(0x7F ^ mask)
	}

	return byte(((seg << 4) | ((v >> (seg + 3)) & 0xF)) ^ mask)
}

func refMuLawToLinear(u byte) int {
	const bias = 0x84

	u = ^u
	t := (int(u&0x0F) << 3) + bias
	t <<= (u & 0x70) >> 4

	if u&0x80 != 0 {
		return bias - t
	}

	return t - bias
}

// TestG711AgainstSunReference checks every code in both directions, and every 16-bit input, against the reference.
func TestG711AgainstSunReference(t *testing.T) {
	var out [1]float64

	for c := range 256 {
		pcm.ALaw{}.Decode(out[:], []byte{byte(c)})
		require.InDelta(t, float64(refALawToLinear(byte(c)))/32768, out[0], 0, "A-law decode %#02x", c)

		pcm.MuLaw{}.Decode(out[:], []byte{byte(c)})
		require.InDelta(t, float64(refMuLawToLinear(byte(c)))/32768, out[0], 0, "µ-law decode %#02x", c)
	}

	for v := math.MinInt16; v <= math.MaxInt16; v++ {
		f := float64(v) / 32768
		require.Equal(t, refLinearToALaw(v), pcm.ALaw{}.AppendEncode(nil, []float64{f})[0], "A-law encode %d", v)
		require.Equal(t, refLinearToMuLaw(v), pcm.MuLaw{}.AppendEncode(nil, []float64{f})[0], "µ-law encode %d", v)
	}
}
