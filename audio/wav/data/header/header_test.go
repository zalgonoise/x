package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/wav/data/header"
)

func TestNew(t *testing.T) {
	t.Run("NewData", func(t *testing.T) {
		wants := &header.Header{Subchunk2ID: header.Data}

		out := header.NewData()

		require.Equal(t, wants, out)
	})

	t.Run("WithDataID", func(t *testing.T) {
		wants := &header.Header{Subchunk2ID: header.Data}

		out := header.New(header.Data)

		require.Equal(t, wants, out)
	})

	t.Run("NewJunk", func(t *testing.T) {
		wants := &header.Header{Subchunk2ID: header.Junk}

		out := header.NewJunk()

		require.Equal(t, wants, out)
	})

	t.Run("WithJunkID", func(t *testing.T) {
		wants := &header.Header{Subchunk2ID: header.Junk}

		out := header.New(header.Junk)

		require.Equal(t, wants, out)
	})

	t.Run("InvalidID", func(t *testing.T) {
		out := header.New([4]byte{0, 1, 2, 3})

		require.Nil(t, out)
	})
}

func TestHeader_Read(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &header.Header{Subchunk2ID: header.Data, Subchunk2Size: 4097}
		wants := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		out := make([]byte, header.Size)

		n, err := h.Read(out)

		require.NoError(t, err)
		require.Equal(t, header.Size, n)
		require.Equal(t, wants, out)
	})
}

func TestHeader_Write(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wants := &header.Header{Subchunk2ID: header.Data, Subchunk2Size: 4097}
		input := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		h := new(header.Header)

		n, err := h.Write(input)

		require.NoError(t, err)
		require.Equal(t, header.Size, n)
		require.Equal(t, wants, h)
	})
}

func TestHeader_Bytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &header.Header{Subchunk2ID: header.Data, Subchunk2Size: 4097}
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
		wants := &header.Header{Subchunk2ID: header.Data, Subchunk2Size: 4097}
		input := []byte{
			100, 97, 116, 97, // "data"
			1, 16, 0, 0, // 4097 little endian
		}

		h, err := header.From(input)

		require.NoError(t, err)
		require.Equal(t, wants, h)
	})
}

func BenchmarkHeader_ReadWrite(b *testing.B) {
	b.Run("Read", func(b *testing.B) {
		h := &header.Header{Subchunk2ID: header.Data, Subchunk2Size: 4097}

		out := make([]byte, header.Size)

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

		h := new(header.Header)
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
