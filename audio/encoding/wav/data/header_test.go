package data_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

func TestNew(t *testing.T) {
	t.Run("NewData", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Data}

		out := data.NewData()

		require.Equal(t, wants, out)
	})

	t.Run("WithDataID", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Data}

		out := data.New(data.Data)

		require.Equal(t, wants, out)
	})

	t.Run("NewJunk", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Junk}

		out := data.NewJunk()

		require.Equal(t, wants, out)
	})

	t.Run("WithJunkID", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Junk}

		out := data.New(data.Junk)

		require.Equal(t, wants, out)
	})

	t.Run("InvalidID", func(t *testing.T) {
		out := data.New([4]byte{0, 1, 2, 3})

		require.Nil(t, out)
	})
}

func TestHeader_Read(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &data.Header{Subchunk2ID: data.Data, Subchunk2Size: 4097}
		wants := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		out := make([]byte, data.Size)

		n, err := h.Read(out)

		require.NoError(t, err)
		require.Equal(t, data.Size, n)
		require.Equal(t, wants, out)
	})
}

func TestHeader_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Data, Subchunk2Size: 4097}
		input := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		h := new(data.Header)

		n, err := h.Write(input)

		require.NoError(t, err)
		require.Equal(t, data.Size, n)
		require.Equal(t, wants, h)
	})
}

func TestHeader_Bytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &data.Header{Subchunk2ID: data.Data, Subchunk2Size: 4097}
		wants := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		out := h.Bytes()

		require.Equal(t, wants, out)
	})
}

func TestFrom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wants := &data.Header{Subchunk2ID: data.Data, Subchunk2Size: 4097}
		input := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		h, err := data.From(input)

		require.NoError(t, err)
		require.Equal(t, wants, h)
	})
}

func BenchmarkHeader_ReadWrite(b *testing.B) {
	b.Run("Read", func(b *testing.B) {
		h := &data.Header{Subchunk2ID: data.Data, Subchunk2Size: 4097}

		out := make([]byte, data.Size)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := h.Read(out)
			if err != nil {
				b.Error(err)
			}
		}
		_ = out
	})

	b.Run("Write", func(b *testing.B) {
		input := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		h := new(data.Header)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := h.Write(input)
			if err != nil {
				b.Error(err)
			}
		}
		_ = h
	})
}
