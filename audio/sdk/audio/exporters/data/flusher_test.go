package data

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/osc"
)

func TestFlusher(t *testing.T) {
	h, err := wav.NewHeader(44100, 16, 2, wav.PCMFormat)
	require.NoError(t, err)

	id := uuid.New().String()

	for _, testcase := range []struct {
		name      string
		outSize   int
		data      []byte
		numWrites int
		wants     []byte
	}{
		{
			name:      "Success/Small",
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 1,
			wants:     []byte{},
		},
		{
			name:      "Success/FilledOnce",
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 2,
			wants:     []byte("gold!gold!"),
		},
		{
			name:      "Success/DoubleRun",
			outSize:   20,
			data:      []byte("gold!"),
			numWrites: 5,
			wants:     []byte("gold!gold!gold!gold!"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			output := make([]byte, 0, testcase.outSize)
			f := NewFlusher(10, func(id string, _ *wav.Header, data []byte) error {
				t.Log(id)
				output = append(output, data...)

				return nil
			})

			f.cap = 10
			f.data = make([]byte, 10)

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

func TestFlusher_Write(t *testing.T) {
	h, err := wav.NewHeader(44100, 16, 1, wav.PCMFormat)
	require.NoError(t, err)

	w, err := wav.FromHeader(h)
	require.NoError(t, err)

	w.Generate(osc.SineWave, 2000, time.Minute)

	stream := w.Data.Bytes()

	t.Log("original:", len(stream))

	t.Run("Uncompressed", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		f := NewFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		_, err := f.Write("test", h, stream)
		require.NoError(t, err)

		t.Log(buf.Len())
	})

	t.Run("ZLib", func(t *testing.T) {

		buf := bytes.NewBuffer(nil)

		f := NewZLibFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		_, err := f.Write("test", h, stream)
		require.NoError(t, err)

		t.Log(buf.Len())
	})

	t.Run("GZip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		f := NewGZipFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		_, err := f.Write("test", h, stream)
		require.NoError(t, err)

		t.Log(buf.Len())
	})
}
