package lpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZigZag(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input int8
		wants uint8
	}{
		{
			name:  "Zero",
			input: 0,
			wants: 1,
		},
		{
			name:  "X1",
			input: 1,
			wants: 2,
		},
		{
			name:  "X-1",
			input: -1,
			wants: 3,
		},
		{
			name:  "X2",
			input: 2,
			wants: 4,
		},
		{
			name:  "X-2",
			input: -2,
			wants: 5,
		},
		{
			name:  "X3",
			input: 3,
			wants: 6,
		},
		{
			name:  "X-3",
			input: -3,
			wants: 7,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, zigZag[uint8](testcase.input))
		})
	}
}
