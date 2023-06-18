package ptr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplit2(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input []int
		wants func() [][2]int // process is destructive, cannot pre-declare a data object ahead of call
	}{
		{
			name:  "SliceOf5",
			input: []int{1, 2, 3, 4, 5},
			wants: func() [][2]int { return [][2]int{{1, 2}, {3, 4}, {5, 0}} },
		},
		{
			name:  "SliceOf1",
			input: []int{1},
			wants: func() [][2]int { return [][2]int{{1, 0}} },
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out := Split2(testcase.input)

			require.Equal(t, testcase.wants(), out)
		})
	}
}

func TestSplit3(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input []int
		wants func() [][3]int // process is destructive, cannot pre-declare a data object ahead of call
	}{
		{
			name:  "SliceOf5",
			input: []int{1, 2, 3, 4, 5},
			wants: func() [][3]int { return [][3]int{{1, 2, 3}, {4, 5, 0}} },
		},
		{
			name:  "SliceOf1",
			input: []int{1},
			wants: func() [][3]int { return [][3]int{{1, 0, 0}} },
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out := Split3(testcase.input)

			require.Equal(t, testcase.wants(), out)
		})
	}
}

func TestSplit4(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input []int
		wants func() [][4]int // process is destructive, cannot pre-declare a data object ahead of call
	}{
		{
			name:  "SliceOf5",
			input: []int{1, 2, 3, 4, 5},
			wants: func() [][4]int { return [][4]int{{1, 2, 3, 4}, {5, 0, 0, 0}} },
		},
		{
			name:  "SliceOf1",
			input: []int{1},
			wants: func() [][4]int { return [][4]int{{1, 0, 0, 0}} },
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out := Split4(testcase.input)

			require.Equal(t, testcase.wants(), out)
		})
	}
}

func TestSplit5(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input []int
		wants func() [][5]int // process is destructive, cannot pre-declare a data object ahead of call
	}{
		{
			name:  "SliceOf6",
			input: []int{1, 2, 3, 4, 5, 6},
			wants: func() [][5]int { return [][5]int{{1, 2, 3, 4, 5}, {6, 0, 0, 0, 0}} },
		},
		{
			name:  "SliceOf1",
			input: []int{1},
			wants: func() [][5]int { return [][5]int{{1, 0, 0, 0, 0}} },
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out := Split5(testcase.input)

			require.Equal(t, testcase.wants(), out)
		})
	}
}
