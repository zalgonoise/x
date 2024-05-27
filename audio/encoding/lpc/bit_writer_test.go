package lpc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBits(t *testing.T) {
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
			wants: []byte{},
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
			wants: []byte{0b10110010, 0b00101110},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			w := NewBitWriter(8)

			w.WriteBits(testcase.bits...)

			require.Equal(t, len(testcase.wants), len(w.buffer))
			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], w.buffer[i],
					"mismatching values",
					fmt.Sprintf("#%d wanted %08b", i, testcase.wants[i]),
					fmt.Sprintf("#%d got %08b", i, w.buffer[i]),
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

		_ = w.buffer
	})
}

func FuzzWriteBits(f *testing.F) {
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

		if len(w.buffer) != 1 {
			t.Fail()

			return
		}

		if w.buffer[0] != n {
			t.Fail()
		}
	})
}
