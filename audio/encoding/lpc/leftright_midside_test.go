package lpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMidSide(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input [2]int16
		wants [2]int16
	}{
		{
			name:  "Zero/NoLeftOrRight",
			input: [2]int16{0, 0},
			wants: [2]int16{0, 0},
		},
		{
			name:  "Zero/NoLeftSide",
			input: [2]int16{0, 10},
			wants: [2]int16{10, -10},
		},
		{
			name:  "Zero/NoRightSide",
			input: [2]int16{10, 0},
			wants: [2]int16{10, 10},
		},
		{
			name:  "Set/Equal",
			input: [2]int16{10, 10},
			wants: [2]int16{20, 0},
		},
		{
			name:  "Set/LeftSideBias",
			input: [2]int16{10, 2},
			wants: [2]int16{12, 8},
		},
		{
			name:  "Set/RightSideBias",
			input: [2]int16{2, 10},
			wants: [2]int16{12, -8},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			mid, side := MidSide(testcase.input[0], testcase.input[1])

			require.Equal(t, testcase.wants[0], mid)
			require.Equal(t, testcase.wants[1], side)
		})
	}
}

func TestLeftRight(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input [2]int16
		wants [2]int16
	}{
		{
			name:  "Zero/NoLeftOrRight",
			input: [2]int16{0, 0},
			wants: [2]int16{0, 0},
		},
		{
			name:  "Zero/NoLeftSide",
			input: [2]int16{10, -10},
			wants: [2]int16{0, 10},
		},
		{
			name:  "Zero/NoRightSide",
			input: [2]int16{10, 10},
			wants: [2]int16{10, 0},
		},
		{
			name:  "Set/Equal",
			input: [2]int16{20, 0},
			wants: [2]int16{10, 10},
		},
		{
			name:  "Set/LeftSideBias",
			input: [2]int16{12, 8},
			wants: [2]int16{10, 2},
		},
		{
			name:  "Set/RightSideBias",
			input: [2]int16{12, -8},
			wants: [2]int16{2, 10},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			left, right := LeftRight(testcase.input[0], testcase.input[1])

			require.Equal(t, testcase.wants[0], left)
			require.Equal(t, testcase.wants[1], right)
		})
	}
}

func TestLeftRight_MidSide_FullCircle(t *testing.T) {
	input := []int16{
		0, 0, 12, 36, 27, 90, 30, 121, 22, 75, 15, 37, 2, 12, -3, -9, -4, -22, -3, -17, 1, -1,
		19, 3, 44, 18, 85, 33, 48, 21, 17, 8, 4, 2, 0, 0,
	}

	midSide := make([]int16, 0, len(input))
	output := make([]int16, 0, len(input))

	for i := 1; i < len(input); i += 2 {
		m, s := MidSide(input[i-1], input[i])

		midSide = append(midSide, m, s)
	}

	for i := 1; i < len(midSide); i += 2 {
		l, r := LeftRight(midSide[i-1], midSide[i])

		output = append(output, l, r)
	}

	require.Equal(t, len(input), len(output))
	for i := range input {
		require.Equal(t, input[i], output[i])
	}
}

func BenchmarkMidSide(b *testing.B) {
	var m, s int16

	for i := 0; i < b.N; i++ {
		m, s = MidSide[int16](27, 90)
	}

	_, _ = m, s
}

func BenchmarkLeftRight(b *testing.B) {
	var l, r int16

	for i := 0; i < b.N; i++ {
		l, r = LeftRight[int16](117, -63)
	}

	_, _ = l, r
}

func FuzzMidSide(f *testing.F) {
	f.Add(int16(0), int16(0))
	f.Add(int16(10), int16(0))
	f.Add(int16(0), int16(10))
	f.Add(int16(5), int16(10))
	f.Add(int16(10), int16(5))
	f.Add(int16(-10), int16(5))
	f.Add(int16(5), int16(-10))

	f.Fuzz(func(t *testing.T, left, right int16) {
		m, s := MidSide(left, right)
		l, r := LeftRight(m, s)

		if left != l {
			t.Fail()
		}

		if right != r {
			t.Fail()
		}
	})
}
