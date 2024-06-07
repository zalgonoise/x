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

func TestAsBits(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input uint8
		wants [8]bool
	}{
		{
			name:  "00000000",
			input: 0,
			wants: [8]bool{false, false, false, false, false, false, false, false},
		},
		{
			name:  "00000001",
			input: 1,
			wants: [8]bool{false, false, false, false, false, false, false, true},
		},
		{
			name:  "00000010",
			input: 2,
			wants: [8]bool{false, false, false, false, false, false, true, false},
		},
		{
			name:  "00000100",
			input: 4,
			wants: [8]bool{false, false, false, false, false, true, false, false},
		},
		{
			name:  "00001000",
			input: 8,
			wants: [8]bool{false, false, false, false, true, false, false, false},
		},
		{
			name:  "00010000",
			input: 16,
			wants: [8]bool{false, false, false, true, false, false, false, false},
		},
		{
			name:  "00100000",
			input: 32,
			wants: [8]bool{false, false, true, false, false, false, false, false},
		},
		{
			name:  "01000000",
			input: 64,
			wants: [8]bool{false, true, false, false, false, false, false, false},
		},
		{
			name:  "10000000",
			input: 128,
			wants: [8]bool{true, false, false, false, false, false, false, false},
		},
		{
			name:  "11111111",
			input: 255,
			wants: [8]bool{true, true, true, true, true, true, true, true},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, asBits(testcase.input))
		})
	}
}

func TestGolombWriter_WriteInt8(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		input       int8
		m           int
		wantsBit    uint8
		wantsBuffer []byte
	}{
		{
			name:     "M0/X0",
			input:    0,
			m:        0,
			wantsBit: 0b10000000, // wrote 1
		},
		{
			name:     "M0/X-1",
			input:    -1,
			m:        0,
			wantsBit: 0b01000000, // wrote 010
		},
		{
			name:     "M0/X1",
			input:    1,
			m:        0,
			wantsBit: 0b01100000, // wrote 011
		},
		{
			name:     "M0/X-2",
			input:    -2,
			m:        0,
			wantsBit: 0b00100000, // wrote 00100
		},
		{
			name:     "M0/X2",
			input:    2,
			m:        0,
			wantsBit: 0b00101000, // wrote 00101
		},
		{
			name:     "M0/X-3",
			input:    -3,
			m:        0,
			wantsBit: 0b00110000, // wrote 00110
		},
		{
			name:     "M0/X3",
			input:    3,
			m:        0,
			wantsBit: 0b00111000, // wrote 00111
		},
		{
			name:     "M0/X-4",
			input:    -4,
			m:        0,
			wantsBit: 0b00010000, // wrote 0001000
		},
		{
			name:     "M1/X0",
			input:    0,
			m:        1,
			wantsBit: 0b10000000, // wrote 10
		},
		{
			name:     "M1/X-1",
			input:    -1,
			m:        1,
			wantsBit: 0b11000000, // wrote 11
		},
		{
			name:     "M1/X1",
			input:    1,
			m:        1,
			wantsBit: 0b01000000, // wrote 0100
		},
		{
			name:     "M1/X-2",
			input:    -2,
			m:        1,
			wantsBit: 0b01010000, // wrote 0101
		},
		{
			name:     "M1/X2",
			input:    2,
			m:        1,
			wantsBit: 0b01100000, // wrote 0110
		},
		{
			name:     "M1/X-3",
			input:    -3,
			m:        1,
			wantsBit: 0b01110000, // wrote 0111
		},
		{
			name:     "M1/X3",
			input:    3,
			m:        1,
			wantsBit: 0b00100000, // wrote 001000
		},
		{
			name:        "M0/X-9",
			input:       -9,
			m:           0,
			wantsBit:    0b00000000, // wrote 000010010
			wantsBuffer: []byte{0b00001001},
		},
		{
			name:     "M3/X-3",
			input:    -3,
			m:        3,
			wantsBit: 0b11010000, // wrote 1101
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			w := &ExpGolombWriter{
				w: NewBitWriter(8),
				m: testcase.m,
			}

			w.WriteInt8(testcase.input)

			require.Equal(t, testcase.wantsBit, w.w.bit)
			require.Equal(t, len(testcase.wantsBuffer), len(w.w.Buffer))
			for i := range testcase.wantsBuffer {
				require.Equal(t, testcase.wantsBuffer[i], w.w.Buffer[i])
			}
		})
	}
}
