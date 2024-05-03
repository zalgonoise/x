package data

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/osc"
)

func BenchmarkFlusher(b *testing.B) {
	h, err := wav.NewHeader(44100, 16, 1, wav.PCMFormat)
	require.NoError(b, err)

	w, err := wav.FromHeader(h)
	require.NoError(b, err)

	w.Generate(osc.SineWave, 2000, time.Minute)

	stream := w.Data.Bytes()

	b.ResetTimer()

	b.Run("Uncompressed", func(b *testing.B) {
		buf := bytes.NewBuffer(nil)

		f := NewFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		for i := 0; i < b.N; i++ {
			_, err := f.Write("test", h, stream)
			require.NoError(b, err)
		}

		_ = f

		b.Log(f.len, buf.Len())
	})

	b.Run("ZLib", func(b *testing.B) {
		buf := bytes.NewBuffer(nil)

		f := NewZLibFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		for i := 0; i < b.N; i++ {
			_, err := f.Write("test", h, stream)
			require.NoError(b, err)
		}

		_ = f

		b.Log(f.len, buf.Len())
	})

	b.Run("GZip", func(b *testing.B) {
		buf := bytes.NewBuffer(nil)

		f := NewGZipFlusher(5*time.Second, func(_ string, _ *wav.Header, data []byte) error {
			_, err := buf.Write(data)

			return err
		})

		for i := 0; i < b.N; i++ {
			_, err := f.Write("test", h, stream)
			require.NoError(b, err)
		}

		_ = f

		b.Log(f.len, buf.Len())
	})
}
