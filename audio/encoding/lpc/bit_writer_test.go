package lpc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitWriter_WriteBits(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		bits  []bool
		wants []byte
	}{
		{
			name:  "ZeroWrites",
			bits:  []bool{},
			wants: []byte{},
		},
		{
			name:  "HalfByte",
			bits:  []bool{true, false, true, true},
			wants: []byte{0b00001011},
		},
		{
			name:  "OneByte",
			bits:  []bool{true, false, true, true, false, false, true, false},
			wants: []byte{0b10110010},
		},
		{
			name: "TwoAndAHalfBytes",
			bits: []bool{
				true, false, true, true, false, false, true, false,
				false, false, true, false, true, true, true, false,
				true, false, false, true,
			},
			wants: []byte{0b10110010, 0b00101110, 0b00001001},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			w := NewBitWriter(8)

			w.WriteBits(testcase.bits...)
			w.Flush()

			require.Equal(t, len(testcase.wants), len(w.Buffer))
			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], w.Buffer[i],
					"mismatching values",
					fmt.Sprintf("#%d wanted %08b", i, testcase.wants[i]),
					fmt.Sprintf("#%d got %08b", i, w.Buffer[i]),
				)
			}
		})
	}
}

func TestBitBuffer_WriteBits(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		bits  []bool
		wants []byte
	}{
		{
			name:  "ZeroWrites",
			bits:  []bool{},
			wants: []byte{},
		},
		{
			name:  "HalfByte",
			bits:  []bool{true, false, true, true},
			wants: []byte{0b00001011},
		},
		{
			name:  "OneByte",
			bits:  []bool{true, false, true, true, false, false, true, false},
			wants: []byte{0b10110010},
		},
		{
			name: "TwoAndAHalfBytes",
			bits: []bool{
				true, false, true, true, false, false, true, false,
				false, false, true, false, true, true, true, false,
				true, false, false, true,
			},
			wants: []byte{0b10110010, 0b00101110, 0b00001001},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			w := NewBitBuffer(8)

			w.WriteBits(testcase.bits...)
			w.Flush()

			require.Equal(t, len(testcase.wants), w.Buffer.Len())
			buf := w.Buffer.Bytes()

			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], buf[i],
					"mismatching values",
					fmt.Sprintf("#%d wanted %08b", i, testcase.wants[i]),
					fmt.Sprintf("#%d got %08b", i, buf[i]),
				)
			}
		})
	}
}

func BenchmarkWriteBits(b *testing.B) {
	b.Run("WriteBits", func(b *testing.B) {
		w := NewBitWriter(1024)

		for i := 0; i < b.N; i++ {
			w.WriteBits(
				true, false, true, true, false, false, true, false,
				false, false, true, false, true, true, true, false,
				true, false, false, true,
			)
		}

		_ = w.Buffer
	})
}

func FuzzBitWriter_WriteBits(f *testing.F) {
	uintAsBits := func(n uint8) []bool {
		if n == 0 {
			return make([]bool, 8)
		}

		valueString := strconv.FormatUint(uint64(n), 2)
		input := make([]bool, 8)
		for i, j := len(valueString)-1, 7; i >= 0; i, j = i-1, j-1 {
			if valueString[i] == '1' {
				input[j] = true
			}
		}

		return input
	}

	f.Add(uint8(255))
	f.Add(uint8(0))
	f.Add(uint8(125))
	f.Add(uint8(13))

	f.Fuzz(func(t *testing.T, n uint8) {
		w := NewBitWriter(8)

		w.WriteBits(uintAsBits(n)...)

		if len(w.Buffer) != 1 {
			t.Fail()

			return
		}

		if w.Buffer[0] != n {
			t.Fail()
		}
	})
}

func FuzzBitBuffer_WriteBits(f *testing.F) {
	uintAsBits := func(n uint8) []bool {
		if n == 0 {
			return make([]bool, 8)
		}

		valueString := strconv.FormatUint(uint64(n), 2)
		input := make([]bool, 8)
		for i, j := len(valueString)-1, 7; i >= 0; i, j = i-1, j-1 {
			if valueString[i] == '1' {
				input[j] = true
			}
		}

		return input
	}

	f.Add(uint8(255))
	f.Add(uint8(0))
	f.Add(uint8(125))
	f.Add(uint8(13))

	f.Fuzz(func(t *testing.T, n uint8) {
		w := NewBitBuffer(8)

		w.WriteBits(uintAsBits(n)...)

		if w.Buffer.Len() != 1 {
			t.Fail()

			return
		}

		buf := w.Buffer.Bytes()
		if len(buf) != 1 {
			t.Fail()

			return
		}

		if buf[0] != n {
			t.Fail()
		}
	})
}
