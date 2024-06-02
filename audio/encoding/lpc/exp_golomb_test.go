package lpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolombEncodeDecode(t *testing.T) {
	for _, testcase := range []struct {
		name string
		x    uint64
		m    uint64
		q    uint64
		r    uint64
	}{
		{
			name: "M0/X3",
			x:    3,
			m:    0,
			q:    2,
			r:    4,
		},
		{
			name: "M0/X8",
			x:    8,
			m:    0,
			q:    3,
			r:    9,
		},
		{
			name: "M3/X14",
			x:    14,
			m:    3,
			q:    1,
			r:    6,
		},
		{
			name: "M1/X14",
			x:    14,
			m:    1,
			q:    3,
			r:    12,
		},
		{
			name: "M10/X1500",
			x:    1500,
			m:    10,
			q:    1,
			r:    476,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			q, r, ok := GolombEncode64(testcase.x, testcase.m)
			require.True(t, ok)
			require.Equal(t, testcase.q, q)
			require.Equal(t, testcase.r, r)
			x, ok := GolombDecode64(testcase.m, testcase.r)
			require.True(t, ok)
			require.Equal(t, testcase.x, x)
		})
	}
}

func FuzzGolombEncodeDecode64(f *testing.F) {
	f.Add(uint64(14), uint64(3))
	f.Add(uint64(14), uint64(1))
	f.Add(uint64(1500), uint64(10))

	f.Fuzz(func(t *testing.T, x, m uint64) {
		_, r, ok := GolombEncode64(x, m)
		if !ok && (m == 0 || m > 63) {
			return
		}

		newX, ok := GolombDecode64(m, r)
		if !ok {
			t.Fail()
		}

		if newX != x {
			t.Fail()
		}
	})
}

func FuzzGolombEncodeDecode32(f *testing.F) {
	f.Add(uint32(14), uint32(3))
	f.Add(uint32(14), uint32(1))
	f.Add(uint32(1500), uint32(10))

	f.Fuzz(func(t *testing.T, x, m uint32) {
		_, r, ok := GolombEncode32(x, m)
		if !ok && (m == 0 || m > 31) {
			return
		}

		newX, ok := GolombDecode32(m, r)
		if !ok {
			t.Fail()
		}

		if newX != x {
			t.Fail()
		}
	})
}

func FuzzGolombEncodeDecode16(f *testing.F) {
	f.Add(uint16(14), uint16(3))
	f.Add(uint16(14), uint16(1))
	f.Add(uint16(1500), uint16(10))

	f.Fuzz(func(t *testing.T, x, m uint16) {
		_, r, ok := GolombEncode16(x, m)
		if !ok && (m == 0 || m > 15) {
			return
		}

		newX, ok := GolombDecode16(m, r)
		if !ok {
			t.Fail()
		}

		if newX != x {
			t.Fail()
		}
	})
}

func FuzzGolombEncodeDecode8(f *testing.F) {
	f.Add(uint8(14), uint8(3))
	f.Add(uint8(14), uint8(1))
	f.Add(uint8(255), uint8(10))

	f.Fuzz(func(t *testing.T, x, m uint8) {
		_, r, ok := GolombEncode8(x, m)
		if !ok && (m == 0 || m > 7) {
			return
		}

		newX, ok := GolombDecode8(m, r)
		if !ok {
			t.Fail()
		}

		if newX != x {
			t.Fail()
		}
	})
}

func TestBitLength(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input uint8
		wants int
	}{
		{
			name:  "00000000",
			input: 0,
			wants: 0,
		},
		{
			name:  "00000001",
			input: 1,
			wants: 1,
		},
		{
			name:  "00000010",
			input: 2,
			wants: 2,
		},
		{
			name:  "00000100",
			input: 4,
			wants: 3,
		},
		{
			name:  "00001000",
			input: 8,
			wants: 4,
		},
		{
			name:  "00010000",
			input: 16,
			wants: 5,
		},
		{
			name:  "00100000",
			input: 32,
			wants: 6,
		},
		{
			name:  "01000000",
			input: 64,
			wants: 7,
		},
		{
			name:  "10000000",
			input: 128,
			wants: 8,
		},
		{
			name:  "11111111",
			input: 255,
			wants: 8,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, bitLength(testcase.input))
		})
	}
}
