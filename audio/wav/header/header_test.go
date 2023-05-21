package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/wav/header"
)

const (
	sampleRate  uint32 = 44100
	bitDepth    uint16 = 16
	numChannels uint16 = 2
)

var (
	defaultChunkID     = [4]byte{82, 73, 70, 70}
	defaultFormat      = [4]byte{87, 65, 86, 69}
	defaultSubchunk1ID = [4]byte{102, 109, 116, 32}
)

func TestHeader_Read(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &header.Header{
			ChunkID:       defaultChunkID,
			ChunkSize:     0,
			Format:        defaultFormat,
			Subchunk1ID:   defaultSubchunk1ID,
			Subchunk1Size: 16,
			AudioFormat:   (uint16)(header.PCMFormat),
			NumChannels:   numChannels,
			SampleRate:    sampleRate,
			ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / 8,
			BlockAlign:    bitDepth * numChannels / 8,
			BitsPerSample: bitDepth,
		}

		wants := []byte{
			82, 73, 70, 70, // ChunkID
			0, 0, 0, 0, // ChunkSize
			87, 65, 86, 69, // Format
			102, 109, 116, 32, // Subchunk1ID
			16, 0, 0, 0, // Subchunk1Size
			1, 0, // AudioFormat
			2, 0, // NumChannels
			68, 172, 0, 0, // SampleRate
			16, 177, 2, 0, // ByteRate
			4, 0, // BlockAlign
			16, 0, // BitsPerSample
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
		input := []byte{
			82, 73, 70, 70, // ChunkID
			0, 0, 0, 0, // ChunkSize
			87, 65, 86, 69, // Format
			102, 109, 116, 32, // Subchunk1ID
			16, 0, 0, 0, // Subchunk1Size
			1, 0, // AudioFormat
			2, 0, // NumChannels
			68, 172, 0, 0, // SampleRate
			16, 177, 2, 0, // ByteRate
			4, 0, // BlockAlign
			16, 0, // BitsPerSample
		}

		wants := &header.Header{
			ChunkID:       defaultChunkID,
			ChunkSize:     0,
			Format:        defaultFormat,
			Subchunk1ID:   defaultSubchunk1ID,
			Subchunk1Size: 16,
			AudioFormat:   (uint16)(header.PCMFormat),
			NumChannels:   numChannels,
			SampleRate:    sampleRate,
			ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / 8,
			BlockAlign:    bitDepth * numChannels / 8,
			BitsPerSample: bitDepth,
		}

		h := new(header.Header)

		n, err := h.Write(input)
		require.NoError(t, err)
		require.Equal(t, header.Size, n)
		require.Equal(t, wants, h)
	})
}

func BenchmarkHeader_ReadWrite(b *testing.B) {
	b.Run("Read", func(b *testing.B) {
		h := &header.Header{
			ChunkID:       defaultChunkID,
			ChunkSize:     0,
			Format:        defaultFormat,
			Subchunk1ID:   defaultSubchunk1ID,
			Subchunk1Size: 16,
			AudioFormat:   (uint16)(header.PCMFormat),
			NumChannels:   numChannels,
			SampleRate:    sampleRate,
			ByteRate:      sampleRate * uint32(bitDepth) * uint32(numChannels) / 8,
			BlockAlign:    bitDepth * numChannels / 8,
			BitsPerSample: bitDepth,
		}

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
			82, 73, 70, 70, // ChunkID
			0, 0, 0, 0, // ChunkSize
			87, 65, 86, 69, // Format
			102, 109, 116, 32, // Subchunk1ID
			16, 0, 0, 0, // Subchunk1Size
			1, 0, // AudioFormat
			2, 0, // NumChannels
			68, 172, 0, 0, // SampleRate
			16, 177, 2, 0, // ByteRate
			4, 0, // BlockAlign
			16, 0, // BitsPerSample
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
