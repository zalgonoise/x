package lpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEliasGammaUint8(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input uint8
		wants []bool
	}{
		{
			name:  "Zero",
			input: 0,
			wants: nil,
		},
		{
			name:  "X1",
			input: 1,
			wants: []bool{true},
		},
		{
			name:  "X2",
			input: 2,
			wants: []bool{false, true, false},
		},
		{
			name:  "X3",
			input: 3,
			wants: []bool{false, true, true},
		},
		{
			name:  "X4",
			input: 4,
			wants: []bool{false, false, true, false, false},
		},
		{
			name:  "X5",
			input: 5,
			wants: []bool{false, false, true, false, true},
		},
		{
			name:  "X6",
			input: 6,
			wants: []bool{false, false, true, true, false},
		},
		{
			name:  "X7",
			input: 7,
			wants: []bool{false, false, true, true, true},
		},
		{
			name:  "X8",
			input: 8,
			wants: []bool{false, false, false, true, false, false, false},
		},
		{
			name:  "X9",
			input: 9,
			wants: []bool{false, false, false, true, false, false, true},
		},
		{
			name:  "X10",
			input: 10,
			wants: []bool{false, false, false, true, false, true, false},
		},
		{
			name:  "X11",
			input: 11,
			wants: []bool{false, false, false, true, false, true, true},
		},
		{
			name:  "X12",
			input: 12,
			wants: []bool{false, false, false, true, true, false, false},
		},
		{
			name:  "X13",
			input: 13,
			wants: []bool{false, false, false, true, true, false, true},
		},
		{
			name:  "X14",
			input: 14,
			wants: []bool{false, false, false, true, true, true, false},
		},
		{
			name:  "X15",
			input: 15,
			wants: []bool{false, false, false, true, true, true, true},
		},
		{
			name:  "X16",
			input: 16,
			wants: []bool{false, false, false, false, true, false, false, false, false},
		},
		{
			name:  "X17",
			input: 17,
			wants: []bool{false, false, false, false, true, false, false, false, true},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			code := EliasGammaUint8(testcase.input)
			if code == nil {
				require.Nil(t, testcase.wants)

				return
			}

			require.Equal(t, len(testcase.wants), len(code))
			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], code[i])
			}
		})
	}
}
