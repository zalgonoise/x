package data

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/audio/encoding/wav"
)

func TestFlusher(t *testing.T) {
	h, err := wav.NewHeader(44100, 16, 2, wav.PCMFormat)
	require.NoError(t, err)

	id := uuid.New().String()

	for _, testcase := range []struct {
		name      string
		size      int
		outSize   int
		data      []byte
		numWrites int
		wants     []byte
	}{
		{
			name:      "Success/Small",
			size:      10,
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 1,
			wants:     []byte{},
		},
		{
			name:      "Success/FilledOnce",
			size:      10,
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 2,
			wants:     []byte("gold!gold!"),
		},
		{
			name:      "Success/DoubleRun",
			size:      10,
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 5,
			wants:     []byte("gold!gold!gold!gold!"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			output := make([]byte, 0, testcase.outSize)
			f := NewFlusher(testcase.size, func(id string, _ *wav.Header, data []byte) error {
				t.Log(id)
				output = append(output, data...)

				return nil
			})

			for i := 0; i < testcase.numWrites; i++ {
				_, err = f.Write(id, h, testcase.data)
				require.NoError(t, err)
			}

			require.Equal(t, len(output), len(testcase.wants))
			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i], output[i])
			}
		})
	}
}
